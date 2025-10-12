package browsingdata

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to create a test Chrome cookies database
func createTestCookieDB(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create cookies table matching Chrome's schema
	createTable := `
	CREATE TABLE cookies (
		name TEXT NOT NULL,
		encrypted_value BLOB NOT NULL,
		host_key TEXT NOT NULL,
		path TEXT NOT NULL,
		creation_utc INTEGER NOT NULL,
		expires_utc INTEGER NOT NULL,
		is_secure INTEGER NOT NULL,
		is_httponly INTEGER NOT NULL,
		has_expires INTEGER NOT NULL,
		is_persistent INTEGER NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		t.Fatalf("failed to create cookies table: %v", err)
	}
}

// Helper function to insert test cookie data
func insertTestCookie(t *testing.T, dbPath, name, host, path string, encryptedValue []byte) {
	t.Helper()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	createTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000000
	expireTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000000

	_, err = db.Exec(
		"INSERT INTO cookies VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		name, encryptedValue, host, path, createTime, expireTime, 1, 1, 1, 1,
	)
	if err != nil {
		t.Fatalf("failed to insert test cookie: %v", err)
	}
}

// Helper function to encrypt data like Chrome does
func encryptLikeChrome(t *testing.T, key, plaintext []byte) []byte {
	t.Helper()

	chromeIV := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}

	// Apply PKCS5 padding
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padtext := append(plaintext, make([]byte, padding)...)
	for i := 0; i < padding; i++ {
		padtext[len(plaintext)+i] = byte(padding)
	}

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	encrypted := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, chromeIV)
	mode.CryptBlocks(encrypted, padtext)

	// Add Chrome's 3-byte prefix
	return append([]byte("v10"), encrypted...)
}

// Helper function to create test login data database
func createTestLoginDB(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create logins table matching Chrome's schema
	createTable := `
	CREATE TABLE logins (
		origin_url TEXT NOT NULL,
		username_value TEXT NOT NULL,
		password_value BLOB NOT NULL,
		date_created INTEGER NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		t.Fatalf("failed to create logins table: %v", err)
	}
}

// Helper function to insert test login data
func insertTestLogin(t *testing.T, dbPath, url, username string, encryptedPassword []byte, created int64) {
	t.Helper()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO logins VALUES (?, ?, ?, ?)",
		url, username, encryptedPassword, created,
	)
	if err != nil {
		t.Fatalf("failed to insert test login: %v", err)
	}
}

func TestGetCookie(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (string, string)
		wantErr   bool
		checkFunc func(t *testing.T, cookies []Cookie)
	}{
		{
			name: "successful cookie extraction",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)

				// Create key and encrypt value
				key := []byte("0123456789abcdef")
				plainValue := []byte("cookie_value_123")
				encryptedValue := encryptLikeChrome(t, key, plainValue)

				insertTestCookie(t, dbPath, "test_cookie", "example.com", "/", encryptedValue)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cookies []Cookie) {
				if len(cookies) != 1 {
					t.Errorf("expected 1 cookie, got %d", len(cookies))
					return
				}
				cookie := cookies[0]
				if cookie.KeyName != "test_cookie" {
					t.Errorf("cookie name = %s, want test_cookie", cookie.KeyName)
				}
				if cookie.Host != "example.com" {
					t.Errorf("cookie host = %s, want example.com", cookie.Host)
				}
			},
		},
		{
			name: "cookie with empty encrypted value",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)

				key := []byte("0123456789abcdef")
				// Insert cookie with empty encrypted value
				insertTestCookie(t, dbPath, "empty_cookie", "example.com", "/", []byte{})

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cookies []Cookie) {
				if len(cookies) != 1 {
					t.Errorf("expected 1 cookie, got %d", len(cookies))
					return
				}
				if cookies[0].Value != "" {
					t.Errorf("expected empty value, got %s", cookies[0].Value)
				}
			},
		},
		{
			name: "cookie with invalid encrypted data",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)

				key := []byte("0123456789abcdef")
				// Insert cookie with invalid encrypted value (will cause decryption error)
				invalidEncrypted := []byte("v10invaliddata")
				insertTestCookie(t, dbPath, "invalid_cookie", "example.com", "/", invalidEncrypted)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cookies []Cookie) {
				if len(cookies) != 1 {
					t.Errorf("expected 1 cookie, got %d", len(cookies))
				}
				// Should still return cookie even if decryption fails
			},
		},
		{
			name: "invalid base64 key",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)
				return "invalid-base64!!!", dbPath
			},
			wantErr: true,
		},
		{
			name: "nonexistent database file",
			setup: func(t *testing.T) (string, string) {
				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, "/nonexistent/path/Cookies"
			},
			wantErr: true,
		},
		{
			name: "empty database",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)
				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cookies []Cookie) {
				if len(cookies) != 0 {
					t.Errorf("expected 0 cookies, got %d", len(cookies))
				}
			},
		},
		{
			name: "database with invalid schema",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")

				// Create a database with wrong schema
				db, err := sql.Open("sqlite3", dbPath)
				if err != nil {
					t.Fatalf("failed to create db: %v", err)
				}
				_, err = db.Exec("CREATE TABLE cookies (id INTEGER)")
				if err != nil {
					t.Fatalf("failed to create table: %v", err)
				}
				db.Close()

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: true,
		},
		{
			name: "database not a valid sqlite file",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")

				// Create a file that's not a valid sqlite database
				if err := os.WriteFile(dbPath, []byte("not a database"), 0644); err != nil {
					t.Fatalf("failed to create invalid db: %v", err)
				}

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: true,
		},
		{
			name: "cookie with corrupted row data",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Cookies")
				createTestCookieDB(t, dbPath)

				// Insert data with wrong types to trigger Scan error
				db, _ := sql.Open("sqlite3", dbPath)
				defer db.Close()
				// This will work but may cause scan issues in some scenarios
				db.Exec("INSERT INTO cookies VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)")

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cookies []Cookie) {
				// Should still return cookies array (may be empty or with partial data)
				t.Logf("Got %d cookies with corrupted data", len(cookies))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base64Key, dbPath := tt.setup(t)
			cookies, err := GetCookie(base64Key, dbPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, cookies)
			}
		})
	}
}

func TestGetLoginData(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (string, string)
		wantErr   bool
		checkFunc func(t *testing.T, logins []loginData)
	}{
		{
			name: "successful login data extraction",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				// Create key and encrypt password
				key := []byte("0123456789abcdef")
				plainPassword := []byte("mypassword123")
				encryptedPassword := encryptLikeChrome(t, key, plainPassword)

				created := time.Now().Unix()
				insertTestLogin(t, dbPath, "https://example.com", "testuser", encryptedPassword, created)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 1 {
					t.Errorf("expected 1 login, got %d", len(logins))
					return
				}
				login := logins[0]
				if login.UserName != "testuser" {
					t.Errorf("username = %s, want testuser", login.UserName)
				}
				if login.LoginUrl != "https://example.com" {
					t.Errorf("url = %s, want https://example.com", login.LoginUrl)
				}
				if login.Password != "mypassword123" {
					t.Errorf("password = %s, want mypassword123", login.Password)
				}
			},
		},
		{
			name: "invalid base64 key",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)
				return "invalid-base64!!!", dbPath
			},
			wantErr: true,
		},
		{
			name: "nonexistent database file",
			setup: func(t *testing.T) (string, string) {
				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, "/nonexistent/path/Login Data"
			},
			wantErr: true,
		},
		{
			name: "empty database",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)
				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 0 {
					t.Errorf("expected 0 logins, got %d", len(logins))
				}
			},
		},
		{
			name: "multiple login entries",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				key := []byte("0123456789abcdef")
				created := time.Now().Unix()

				for i := 0; i < 3; i++ {
					password := []byte(fmt.Sprintf("password%d", i))
					encrypted := encryptLikeChrome(t, key, password)
					insertTestLogin(t, dbPath,
						fmt.Sprintf("https://site%d.com", i),
						fmt.Sprintf("user%d", i),
						encrypted,
						created)
				}

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 3 {
					t.Errorf("expected 3 logins, got %d", len(logins))
				}
			},
		},
		{
			name: "login with empty password",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				key := []byte("0123456789abcdef")
				created := time.Now().Unix()
				// Insert login with empty password
				insertTestLogin(t, dbPath, "https://example.com", "testuser", []byte{}, created)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 1 {
					t.Errorf("expected 1 login, got %d", len(logins))
					return
				}
				if logins[0].Password != "" {
					t.Errorf("expected empty password, got %s", logins[0].Password)
				}
			},
		},
		{
			name: "login with invalid encrypted password",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				key := []byte("0123456789abcdef")
				created := time.Now().Unix()
				// Insert login with invalid encrypted password
				invalidEncrypted := []byte("v10invaliddata")
				insertTestLogin(t, dbPath, "https://example.com", "testuser", invalidEncrypted, created)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 1 {
					t.Errorf("expected 1 login, got %d", len(logins))
				}
				// Should still return login even if decryption fails
			},
		},
		{
			name: "login with future timestamp (TimeEpoch path)",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				key := []byte("0123456789abcdef")
				plainPassword := []byte("mypassword123")
				encryptedPassword := encryptLikeChrome(t, key, plainPassword)

				// Use a future timestamp to trigger TimeEpoch path
				futureTimestamp := time.Now().Unix() + 1000000000
				insertTestLogin(t, dbPath, "https://example.com", "testuser", encryptedPassword, futureTimestamp)

				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				if len(logins) != 1 {
					t.Errorf("expected 1 login, got %d", len(logins))
				}
			},
		},
		{
			name: "database with invalid schema for login",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")

				// Create a database with wrong schema
				db, err := sql.Open("sqlite3", dbPath)
				if err != nil {
					t.Fatalf("failed to create db: %v", err)
				}
				_, err = db.Exec("CREATE TABLE logins (id INTEGER)")
				if err != nil {
					t.Fatalf("failed to create table: %v", err)
				}
				db.Close()

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: true,
		},
		{
			name: "login database not a valid sqlite file",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")

				// Create a file that's not a valid sqlite database
				if err := os.WriteFile(dbPath, []byte("not a database"), 0644); err != nil {
					t.Fatalf("failed to create invalid db: %v", err)
				}

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: true,
		},
		{
			name: "login with corrupted row data",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "Login Data")
				createTestLoginDB(t, dbPath)

				// Insert data with wrong types to trigger Scan error
				db, _ := sql.Open("sqlite3", dbPath)
				defer db.Close()
				db.Exec("INSERT INTO logins VALUES (NULL, NULL, NULL, NULL)")

				key := []byte("0123456789abcdef")
				base64Key := base64.StdEncoding.EncodeToString(key)
				return base64Key, dbPath
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logins []loginData) {
				// Should still return logins array
				t.Logf("Got %d logins with corrupted data", len(logins))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base64Key, dbPath := tt.setup(t)
			logins, err := GetLoginData(base64Key, dbPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoginData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, logins)
			}
		})
	}
}

func TestCookieStruct(t *testing.T) {
	// Test Cookie struct field types
	cookie := Cookie{
		Host:         "example.com",
		Path:         "/",
		KeyName:      "test",
		encryptValue: []byte("encrypted"),
		Value:        "decrypted",
		IsSecure:     true,
		IsHTTPOnly:   true,
		HasExpire:    true,
		IsPersistent: true,
		CreateDate:   time.Now(),
		ExpireDate:   time.Now().Add(24 * time.Hour),
	}

	if cookie.Host != "example.com" {
		t.Errorf("Host = %s, want example.com", cookie.Host)
	}
	if !cookie.IsSecure {
		t.Error("IsSecure should be true")
	}
}

func TestLoginDataStruct(t *testing.T) {
	// Test loginData struct field types
	login := loginData{
		LoginUrl:    "https://example.com",
		UserName:    "testuser",
		encryptPass: []byte("encrypted"),
		encryptUser: []byte("encrypted_user"),
		Password:    "mypassword",
		CreateDate:  time.Now(),
	}

	if login.UserName != "testuser" {
		t.Errorf("UserName = %s, want testuser", login.UserName)
	}
	if login.Password != "mypassword" {
		t.Errorf("Password = %s, want mypassword", login.Password)
	}
}

func BenchmarkGetCookie(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "Cookies")
	createTestCookieDB(&testing.T{}, dbPath)

	key := []byte("0123456789abcdef")
	plainValue := []byte("cookie_value_123")
	encryptedValue := encryptLikeChrome(&testing.T{}, key, plainValue)
	insertTestCookie(&testing.T{}, dbPath, "test_cookie", "example.com", "/", encryptedValue)

	base64Key := base64.StdEncoding.EncodeToString(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy the database for each iteration to avoid conflicts
		testPath := filepath.Join(tmpDir, fmt.Sprintf("Cookies_%d", i))
		data, _ := os.ReadFile(dbPath)
		_ = os.WriteFile(testPath, data, 0644)
		_, _ = GetCookie(base64Key, testPath)
		_ = os.Remove(testPath)
	}
}

func BenchmarkGetLoginData(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "Login Data")
	createTestLoginDB(&testing.T{}, dbPath)

	key := []byte("0123456789abcdef")
	plainPassword := []byte("mypassword123")
	encryptedPassword := encryptLikeChrome(&testing.T{}, key, plainPassword)
	created := time.Now().Unix()
	insertTestLogin(&testing.T{}, dbPath, "https://example.com", "testuser", encryptedPassword, created)

	base64Key := base64.StdEncoding.EncodeToString(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy the database for each iteration
		testPath := filepath.Join(tmpDir, fmt.Sprintf("Login_Data_%d", i))
		data, _ := os.ReadFile(dbPath)
		_ = os.WriteFile(testPath, data, 0644)
		_, _ = GetLoginData(base64Key, testPath)
		_ = os.Remove(testPath)
	}
}
