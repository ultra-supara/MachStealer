package decrypter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

func TestAes128CBCDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (key, iv, encrypted []byte, expected []byte)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "successful decryption",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := []byte("0123456789abcdef") // 16 bytes
				iv := []byte("0123456789abcdef")  // 16 bytes
				plaintext := []byte("test message for encryption")

				// Apply PKCS5 padding
				blockSize := aes.BlockSize
				padding := blockSize - len(plaintext)%blockSize
				padtext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)

				// Encrypt
				block, _ := aes.NewCipher(key)
				encrypted := make([]byte, len(padtext))
				mode := cipher.NewCBCEncrypter(block, iv)
				mode.CryptBlocks(encrypted, padtext)

				return key, iv, encrypted, plaintext
			},
			wantErr: false,
		},
		{
			name: "invalid key size",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := []byte("short") // Invalid key size
				iv := []byte("0123456789abcdef")
				encrypted := []byte("some encrypted data")
				return key, iv, encrypted, nil
			},
			wantErr: true,
		},
		{
			name: "encrypted data shorter than block size",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := []byte("0123456789abcdef")
				iv := []byte("0123456789abcdef")
				encrypted := []byte("short") // Less than 16 bytes
				return key, iv, encrypted, nil
			},
			wantErr:   true,
			errSubstr: "less than block size",
		},
		{
			name: "empty encrypted data",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := []byte("0123456789abcdef")
				iv := []byte("0123456789abcdef")
				encrypted := []byte{}
				return key, iv, encrypted, nil
			},
			wantErr:   true,
			errSubstr: "less than block size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, iv, encrypted, expected := tt.setup()
			result, err := aes128CBCDecrypt(key, iv, encrypted)

			if (err != nil) != tt.wantErr {
				t.Errorf("aes128CBCDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errSubstr != "" && err != nil {
				if err.Error() != tt.errSubstr && !bytes.Contains([]byte(err.Error()), []byte(tt.errSubstr)) {
					t.Errorf("aes128CBCDecrypt() error = %v, want error containing %q", err, tt.errSubstr)
				}
			}

			if !tt.wantErr && expected != nil {
				if !bytes.Equal(result, expected) {
					t.Errorf("aes128CBCDecrypt() result = %q, want %q", result, expected)
				}
			}
		})
	}
}

func TestAesGCMDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (encrypted, key, nonce []byte, expected []byte)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "successful decryption",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := make([]byte, 32) // AES-256
				nonce := make([]byte, 12)
				plaintext := []byte("test message for GCM encryption")

				_, _ = rand.Read(key)
				_, _ = rand.Read(nonce)

				// Encrypt using AES-GCM
				block, _ := aes.NewCipher(key)
				gcm, _ := cipher.NewGCM(block)
				encrypted := gcm.Seal(nil, nonce, plaintext, nil)

				return encrypted, key, nonce, plaintext
			},
			wantErr: false,
		},
		{
			name: "invalid key size",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := []byte("short")
				nonce := make([]byte, 12)
				encrypted := []byte("some encrypted data")
				return encrypted, key, nonce, nil
			},
			wantErr: true,
		},
		{
			name: "tampered ciphertext",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := make([]byte, 32)
				nonce := make([]byte, 12)
				plaintext := []byte("test message")

				_, _ = rand.Read(key)
				_, _ = rand.Read(nonce)

				// Encrypt
				block, _ := aes.NewCipher(key)
				gcm, _ := cipher.NewGCM(block)
				encrypted := gcm.Seal(nil, nonce, plaintext, nil)

				// Tamper with the ciphertext
				if len(encrypted) > 0 {
					encrypted[0] ^= 0xFF
				}

				return encrypted, key, nonce, nil
			},
			wantErr: true,
		},
		{
			name: "empty encrypted data",
			setup: func() ([]byte, []byte, []byte, []byte) {
				key := make([]byte, 32)
				nonce := make([]byte, 12)
				_, _ = rand.Read(key)
				_, _ = rand.Read(nonce)
				return []byte{}, key, nonce, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, key, nonce, expected := tt.setup()
			result, err := aesGCMDecrypt(encrypted, key, nonce)

			if (err != nil) != tt.wantErr {
				t.Errorf("aesGCMDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && expected != nil {
				if !bytes.Equal(result, expected) {
					t.Errorf("aesGCMDecrypt() result = %q, want %q", result, expected)
				}
			}
		})
	}
}

func TestPkcs5UnPadding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		blockSize int
		want      []byte
	}{
		{
			name:      "valid padding (1 byte)",
			input:     []byte("hello world\x01"),
			blockSize: 16,
			want:      []byte("hello world"),
		},
		{
			name:      "valid padding (4 bytes)",
			input:     []byte("test\x04\x04\x04\x04"),
			blockSize: 16,
			want:      []byte("test"),
		},
		{
			name:      "full block padding",
			input:     append([]byte("0123456789abcdef"), bytes.Repeat([]byte{16}, 16)...),
			blockSize: 16,
			want:      []byte("0123456789abcdef"),
		},
		{
			name:      "invalid padding (larger than length)",
			input:     []byte("abc\x10"),
			blockSize: 16,
			want:      []byte("abc\x10"), // Should return unchanged
		},
		{
			name:      "invalid padding (larger than block size)",
			input:     []byte("test\x20"),
			blockSize: 16,
			want:      []byte("test\x20"), // Should return unchanged
		},
		{
			name:      "single byte input",
			input:     []byte{0x01},
			blockSize: 16,
			want:      []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkcs5UnPadding(tt.input, tt.blockSize)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("pkcs5UnPadding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkAes128CBCDecrypt(b *testing.B) {
	key := []byte("0123456789abcdef")
	iv := []byte("0123456789abcdef")
	plaintext := []byte("test message for encryption benchmark")

	// Prepare encrypted data
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padtext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, _ := aes.NewCipher(key)
	encrypted := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, padtext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aes128CBCDecrypt(key, iv, encrypted)
	}
}

func BenchmarkAesGCMDecrypt(b *testing.B) {
	key := make([]byte, 32)
	nonce := make([]byte, 12)
	plaintext := []byte("test message for GCM encryption benchmark")

	_, _ = rand.Read(key)
	_, _ = rand.Read(nonce)

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	encrypted := gcm.Seal(nil, nonce, plaintext, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aesGCMDecrypt(encrypted, key, nonce)
	}
}

func BenchmarkPkcs5UnPadding(b *testing.B) {
	input := append([]byte("0123456789abcdef"), bytes.Repeat([]byte{16}, 16)...)
	blockSize := 16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pkcs5UnPadding(input, blockSize)
	}
}
