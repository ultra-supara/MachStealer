package decrypter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

func TestChromium(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (key, encryptPass []byte, expected []byte)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "successful decryption with Chrome encryption",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte("0123456789abcdef") // 16 bytes AES key
				plaintext := []byte("password123")

				// Chrome uses 3-byte prefix + encrypted data
				chromeIV := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}

				// Apply PKCS5 padding
				blockSize := aes.BlockSize
				padding := blockSize - len(plaintext)%blockSize
				padtext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)

				// Encrypt using CBC
				block, _ := aes.NewCipher(key)
				encrypted := make([]byte, len(padtext))
				mode := cipher.NewCBCEncrypter(block, chromeIV)
				mode.CryptBlocks(encrypted, padtext)

				// Add Chrome's 3-byte prefix (v10, v11, etc.)
				chromeEncrypted := append([]byte("v10"), encrypted...)

				return key, chromeEncrypted, plaintext
			},
			wantErr: false,
		},
		{
			name: "empty key error",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte{}
				encryptPass := []byte("v10someencrypteddata1234567890")
				return key, encryptPass, nil
			},
			wantErr:   true,
			errSubstr: "password is empty",
		},
		{
			name: "encrypted data too short (less than 3 bytes)",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte("0123456789abcdef")
				encryptPass := []byte("ab") // Only 2 bytes
				return key, encryptPass, nil
			},
			wantErr:   true,
			errSubstr: "decryption failed",
		},
		{
			name: "encrypted data exactly 3 bytes (no actual encrypted data)",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte("0123456789abcdef")
				encryptPass := []byte("v10") // Only prefix, no encrypted data
				return key, encryptPass, nil
			},
			wantErr:   true,
			errSubstr: "decryption failed",
		},
		{
			name: "nil encrypted data",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte("0123456789abcdef")
				encryptPass := []byte(nil)
				return key, encryptPass, nil
			},
			wantErr:   true,
			errSubstr: "decryption failed",
		},
		{
			name: "valid length but invalid key",
			setup: func() ([]byte, []byte, []byte) {
				key := []byte("wrongkey")
				encryptPass := []byte("v10someencrypteddata1234567890")
				return key, encryptPass, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, encryptPass, expected := tt.setup()
			result, err := Chromium(key, encryptPass)

			if (err != nil) != tt.wantErr {
				t.Errorf("Chromium() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errSubstr != "" && err != nil {
				if err.Error() != tt.errSubstr && !bytes.Contains([]byte(err.Error()), []byte(tt.errSubstr)) {
					t.Errorf("Chromium() error = %v, want error containing %q", err, tt.errSubstr)
				}
			}

			if !tt.wantErr && expected != nil {
				if !bytes.Equal(result, expected) {
					t.Errorf("Chromium() result = %q, want %q", result, expected)
				}
			}
		})
	}
}

func TestDPApi(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "any data returns nil",
			data: []byte("test data"),
			want: nil,
		},
		{
			name: "empty data returns nil",
			data: []byte{},
			want: nil,
		},
		{
			name: "nil data returns nil",
			data: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DPApi(tt.data)
			if err != nil {
				t.Errorf("DPApi() unexpected error = %v", err)
				return
			}
			if got != nil {
				t.Errorf("DPApi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChromiumRoundTrip(t *testing.T) {
	// Test complete encryption -> decryption cycle
	key := []byte("0123456789abcdef")
	plaintext := []byte("test password with special chars: !@#$%^&*()")

	// Encrypt
	chromeIV := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padtext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	encrypted := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, chromeIV)
	mode.CryptBlocks(encrypted, padtext)

	// Add Chrome prefix
	chromeEncrypted := append([]byte("v10"), encrypted...)

	// Decrypt
	result, err := Chromium(key, chromeEncrypted)
	if err != nil {
		t.Fatalf("Chromium() error = %v", err)
	}

	if !bytes.Equal(result, plaintext) {
		t.Errorf("Round trip failed: got %q, want %q", result, plaintext)
	}
}

func BenchmarkChromium(b *testing.B) {
	key := []byte("0123456789abcdef")
	plaintext := []byte("benchmark password")

	// Prepare encrypted data
	chromeIV := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padtext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, _ := aes.NewCipher(key)
	encrypted := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, chromeIV)
	mode.CryptBlocks(encrypted, padtext)
	chromeEncrypted := append([]byte("v10"), encrypted...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Chromium(key, chromeEncrypted)
	}
}

func BenchmarkDPApi(b *testing.B) {
	data := []byte("test data for DPAPI benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DPApi(data)
	}
}
