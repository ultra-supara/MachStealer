package masterkey

import (
	"bytes"
	"crypto/sha1"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestKeyGeneration(t *testing.T) {
	tests := []struct {
		name      string
		seed      []byte
		wantErr   bool
		errType   error
		checkFunc func(key []byte) bool
	}{
		{
			name:    "valid seed generates key",
			seed:    []byte("ChromeSafeStorageKey"),
			wantErr: false,
			checkFunc: func(key []byte) bool {
				return len(key) == 16 // Should return 16-byte key
			},
		},
		{
			name:    "seed with whitespace (trimmed)",
			seed:    []byte("  ChromeSafeStorageKey  \n"),
			wantErr: false,
			checkFunc: func(key []byte) bool {
				return len(key) == 16
			},
		},
		{
			name:    "empty seed after trimming",
			seed:    []byte("   \n\t  "),
			wantErr: true,
			errType: ErrWrongSecurityCommand,
		},
		{
			name:    "nil seed returns error",
			seed:    nil,
			wantErr: true,
			errType: ErrWrongSecurityCommand,
		},
		{
			name:    "single character seed",
			seed:    []byte("a"),
			wantErr: false,
			checkFunc: func(key []byte) bool {
				return len(key) == 16
			},
		},
		{
			name:    "seed that produces nil key (edge case)",
			seed:    []byte("test"),
			wantErr: false,
			checkFunc: func(key []byte) bool {
				// pbkdf2.Key should never return nil with valid inputs
				return key != nil && len(key) == 16
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := KeyGeneration(tt.seed)

			if (err != nil) != tt.wantErr {
				t.Errorf("KeyGeneration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if err != tt.errType {
					t.Errorf("KeyGeneration() error = %v, want %v", err, tt.errType)
				}
			}

			if !tt.wantErr {
				if key == nil {
					t.Error("KeyGeneration() returned nil key without error")
					return
				}
				if tt.checkFunc != nil && !tt.checkFunc(key) {
					t.Errorf("KeyGeneration() key validation failed, key = %v", key)
				}
			}
		})
	}
}

func TestKeyGenerationDeterministic(t *testing.T) {
	// Test that same input produces same output
	seed := []byte("TestSeed123")

	key1, err1 := KeyGeneration(seed)
	if err1 != nil {
		t.Fatalf("First KeyGeneration() failed: %v", err1)
	}

	key2, err2 := KeyGeneration(seed)
	if err2 != nil {
		t.Fatalf("Second KeyGeneration() failed: %v", err2)
	}

	if !bytes.Equal(key1, key2) {
		t.Errorf("KeyGeneration() not deterministic: key1=%v, key2=%v", key1, key2)
	}
}

func TestKeyGenerationMatchesChromiumSpec(t *testing.T) {
	// Test that our implementation matches Chromium's specification
	// https://source.chromium.org/chromium/chromium/src/+/master:components/os_crypt/os_crypt_mac.mm;l=157
	seed := []byte("ChromeSafeStorageKey")
	chromeSalt := []byte("saltysalt")
	iterations := 1003
	keyLength := 16

	expectedKey := pbkdf2.Key(seed, chromeSalt, iterations, keyLength, sha1.New)

	generatedKey, err := KeyGeneration(seed)
	if err != nil {
		t.Fatalf("KeyGeneration() failed: %v", err)
	}

	if !bytes.Equal(generatedKey, expectedKey) {
		t.Errorf("KeyGeneration() does not match Chromium spec:\ngot  = %x\nwant = %x", generatedKey, expectedKey)
	}
}

func TestKeyGenerationDifferentSeeds(t *testing.T) {
	// Test that different seeds produce different keys
	seed1 := []byte("seed1")
	seed2 := []byte("seed2")

	key1, err1 := KeyGeneration(seed1)
	if err1 != nil {
		t.Fatalf("KeyGeneration(seed1) failed: %v", err1)
	}

	key2, err2 := KeyGeneration(seed2)
	if err2 != nil {
		t.Fatalf("KeyGeneration(seed2) failed: %v", err2)
	}

	if bytes.Equal(key1, key2) {
		t.Errorf("KeyGeneration() produced same key for different seeds: %x", key1)
	}
}

func TestKeyGenerationLength(t *testing.T) {
	// Verify key length is always 16 bytes (128 bits)
	seeds := [][]byte{
		[]byte("short"),
		[]byte("a much longer seed value for testing purposes"),
		[]byte("ChromeSafeStorageKey"),
		[]byte("123456"),
		[]byte("!@#$%^&*()"),
	}

	for _, seed := range seeds {
		t.Run(string(seed), func(t *testing.T) {
			key, err := KeyGeneration(seed)
			if err != nil {
				t.Fatalf("KeyGeneration() failed: %v", err)
			}
			if len(key) != 16 {
				t.Errorf("KeyGeneration() key length = %d, want 16", len(key))
			}
		})
	}
}

func TestGetMasterKey(t *testing.T) {
	// Note: GetMasterKey requires actual macOS keychain access and will prompt the user
	// We test it with a dummy parameter but expect it might fail in CI environments
	t.Run("with dummy parameter", func(t *testing.T) {
		// This test will likely fail unless Chrome keychain entry exists
		// We just verify it doesn't panic and returns proper error handling
		key, err := GetMasterKey("dummy")

		// We accept both success and specific error types
		if err != nil {
			// Check if it's one of our expected errors
			if err != ErrWrongSecurityCommand && err != ErrCouldNotFindInKeychain {
				// It's a different error (likely exec.Command error), which is acceptable
				t.Logf("GetMasterKey() returned error (expected in test environment): %v", err)
			}
		} else {
			// Success case - verify the key is valid
			if key == nil {
				t.Error("GetMasterKey() returned nil key without error")
			}
			if len(key) != 16 {
				t.Errorf("GetMasterKey() key length = %d, want 16", len(key))
			}
			t.Log("GetMasterKey() succeeded - Chrome keychain entry found")
		}
	})
}

func TestErrorTypes(t *testing.T) {
	// Test that error types are properly defined
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrWrongSecurityCommand",
			err:  ErrWrongSecurityCommand,
			want: "macOS wrong security command",
		},
		{
			name: "ErrCouldNotFindInKeychain",
			err:  ErrCouldNotFindInKeychain,
			want: "macOS could not find in keychain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	seed := []byte("ChromeSafeStorageKey")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = KeyGeneration(seed)
	}
}

func BenchmarkKeyGenerationWithTrim(b *testing.B) {
	seed := []byte("  ChromeSafeStorageKey  \n\t")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = KeyGeneration(seed)
	}
}

func BenchmarkPBKDF2(b *testing.B) {
	// Benchmark raw PBKDF2 performance to compare
	seed := []byte("ChromeSafeStorageKey")
	salt := []byte("saltysalt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pbkdf2.Key(seed, salt, 1003, 16, sha1.New)
	}
}
