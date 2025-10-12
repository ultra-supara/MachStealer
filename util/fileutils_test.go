package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileCopy(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (string, string)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "successful copy",
			setup: func() (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				content := []byte("test content for file copy")
				if err := os.WriteFile(src, content, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return src, dst
			},
			wantErr: false,
		},
		{
			name: "source file does not exist",
			setup: func() (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "nonexistent.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				return src, dst
			},
			wantErr:   true,
			errSubstr: "FileCopy failed",
		},
		{
			name: "destination directory does not exist",
			setup: func() (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.txt")
				content := []byte("test content")
				if err := os.WriteFile(src, content, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				dst := filepath.Join(tmpDir, "nonexistent_dir", "dest.txt")
				return src, dst
			},
			wantErr:   true,
			errSubstr: "FileCopy failed",
		},
		{
			name: "copy empty file",
			setup: func() (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "empty.txt")
				dst := filepath.Join(tmpDir, "empty_copy.txt")
				if err := os.WriteFile(src, []byte{}, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return src, dst
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup()
			err := FileCopy(src, dst)

			if (err != nil) != tt.wantErr {
				t.Errorf("FileCopy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errSubstr != "" && err != nil {
				if err.Error() == "" {
					t.Errorf("FileCopy() expected error containing %q, got empty error", tt.errSubstr)
				}
			}

			if !tt.wantErr {
				// Verify content was copied correctly
				srcContent, err := os.ReadFile(src)
				if err != nil {
					t.Fatalf("failed to read source: %v", err)
				}
				dstContent, err := os.ReadFile(dst)
				if err != nil {
					t.Fatalf("failed to read destination: %v", err)
				}
				if string(srcContent) != string(dstContent) {
					t.Errorf("content mismatch: src=%q, dst=%q", srcContent, dstContent)
				}
			}
		})
	}
}

func TestTimeStamp(t *testing.T) {
	tests := []struct {
		name      string
		stamp     int64
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "normal timestamp (2020-01-01)",
			stamp:     1577836800,
			wantYear:  2020,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			name:      "zero timestamp (epoch)",
			stamp:     0,
			wantYear:  1970,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			name:      "negative timestamp",
			stamp:     -86400,
			wantYear:  1969,
			wantMonth: time.December,
			wantDay:   31,
		},
		{
			name:      "very large timestamp (>9999 year)",
			stamp:     253402300800,
			wantYear:  9999,
			wantMonth: time.December,
			wantDay:   13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeStamp(tt.stamp)
			if result.Year() != tt.wantYear {
				t.Errorf("TimeStamp() year = %d, want %d", result.Year(), tt.wantYear)
			}
			if result.Month() != tt.wantMonth {
				t.Errorf("TimeStamp() month = %v, want %v", result.Month(), tt.wantMonth)
			}
			if result.Day() != tt.wantDay {
				t.Errorf("TimeStamp() day = %d, want %d", result.Day(), tt.wantDay)
			}
		})
	}
}

func TestTimeEpoch(t *testing.T) {
	tests := []struct {
		name      string
		epoch     int64
		wantYear  int
		checkYear bool
	}{
		{
			name:      "normal epoch time",
			epoch:     13268320000000000,
			wantYear:  2021,
			checkYear: true,
		},
		{
			name:      "zero epoch",
			epoch:     0,
			wantYear:  1601,
			checkYear: true,
		},
		{
			name:      "epoch exceeds max (returns 2049)",
			epoch:     99633311750000000,
			wantYear:  2049,
			checkYear: true,
		},
		{
			name:      "small epoch value",
			epoch:     10000000000000,
			wantYear:  1601,
			checkYear: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeEpoch(tt.epoch)
			if tt.checkYear && result.Year() != tt.wantYear {
				t.Errorf("TimeEpoch() year = %d, want %d", result.Year(), tt.wantYear)
			}
		})
	}
}

func TestIntToBool(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  bool
	}{
		{
			name:  "zero returns false",
			input: 0,
			want:  false,
		},
		{
			name:  "negative one returns false",
			input: -1,
			want:  false,
		},
		{
			name:  "positive number returns true",
			input: 1,
			want:  true,
		},
		{
			name:  "large positive number returns true",
			input: 999,
			want:  true,
		},
		{
			name:  "negative number (not -1) returns true",
			input: -5,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntToBool(tt.input)
			if got != tt.want {
				t.Errorf("IntToBool(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIntToBoolInt64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  bool
	}{
		{
			name:  "zero returns false",
			input: 0,
			want:  false,
		},
		{
			name:  "negative one returns false",
			input: -1,
			want:  false,
		},
		{
			name:  "positive number returns true",
			input: 1,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntToBool(tt.input)
			if got != tt.want {
				t.Errorf("IntToBool(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkFileCopy(b *testing.B) {
	tmpDir := b.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	content := make([]byte, 1024*1024) // 1MB file
	if err := os.WriteFile(src, content, 0644); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := filepath.Join(tmpDir, "dest_"+string(rune(i))+".txt")
		_ = FileCopy(src, dst)
		_ = os.Remove(dst)
	}
}

func BenchmarkTimeStamp(b *testing.B) {
	stamp := int64(1577836800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TimeStamp(stamp)
	}
}

func BenchmarkTimeEpoch(b *testing.B) {
	epoch := int64(13268320000000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TimeEpoch(epoch)
	}
}

func BenchmarkIntToBool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IntToBool(1)
	}
}
