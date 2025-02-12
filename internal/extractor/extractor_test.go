package extractor

import (
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func createTestFile(t *testing.T, content string) (string, func()) {
	tmpfile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name(), func() {
		os.Remove(tmpfile.Name())
	}
}

func TestExtractor_Extract(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   Config
		want     Results
		wantErr  bool
		setupCtx func() (context.Context, context.CancelFunc)
	}{
		{
			name: "all patterns",
			input: `https://example.com/users?id=123&token=abc
user@example.com
192.168.1.1
550e8400-e29b-41d4-a716-446655440000`,
			config: Config{
				UUIDVersion:    4,
				ExtractEmails:  true,
				ExtractDomains: true,
				ExtractIPs:     true,
				ExtractParams:  true,
			},
			want: Results{
				UUIDs: map[string]bool{
					"550e8400-e29b-41d4-a716-446655440000": true,
				},
				Emails: map[string]bool{
					"user@example.com": true,
				},
				Domains: map[string]bool{
					"example.com": true,
				},
				IPs: map[string]bool{
					"192.168.1.1": true,
				},
				Params: map[string]bool{
					"id=123":    true,
					"token=abc": true,
				},
			},
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
		{
			name:  "timeout context",
			input: "very large file simulation\n" + strings.Repeat("a", 1000000),
			config: Config{
				ExtractEmails: true,
			},
			wantErr: true,
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Nanosecond)
			},
		},
		{
			name: "invalid IP addresses",
			input: `256.256.256.256
192.168.1.1
999.0.0.1`,
			config: Config{
				ExtractIPs: true,
			},
			want: Results{
				IPs: map[string]bool{
					"192.168.1.1": true,
				},
			},
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
		{
			name: "multiple UUID versions",
			input: `550e8400-e29b-41d4-a716-446655440000
550e8400-e29b-11d4-a716-446655440000`,
			config: Config{
				UUIDVersion: 1,
			},
			want: Results{
				UUIDs: map[string]bool{
					"550e8400-e29b-11d4-a716-446655440000": true,
				},
			},
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupCtx()
			defer cancel()

			ext, err := New(tt.config)
			if err != nil {
				t.Fatalf("Failed to create extractor: %v", err)
			}

			reader := strings.NewReader(tt.input)
			got, err := ext.Extract(ctx, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractor_ExtractWithLargeFile(t *testing.T) {
	largeContent := strings.Repeat("test content\n", 1024*1024*11) // Over 100MB
	filepath, cleanup := createTestFile(t, largeContent)
	defer cleanup()

	ext, err := New(Config{ExtractEmails: true})
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	file, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	ctx := context.Background()
	_, err = ext.Extract(ctx, file)
	if err == nil || !strings.Contains(err.Error(), "file too large") {
		t.Errorf("Expected 'file too large' error, got %v", err)
	}
}

func TestExtractor_ExtractWithInvalidFile(t *testing.T) {
	ext, err := New(Config{ExtractEmails: true})
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	_, err = ext.Extract(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil reader, got nil")
	}

	failingReader := &failingReader{}
	_, err = ext.Extract(context.Background(), failingReader)
	if err == nil {
		t.Error("Expected error for failing reader, got nil")
	}
}

// failingReader implements io.Reader for testing error cases
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestExtractor_ExtractWithErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (io.Reader, Config)
		wantErr string
	}{
		{
			name: "nil reader",
			setup: func() (io.Reader, Config) {
				return nil, Config{}
			},
			wantErr: "nil reader",
		},
		{
			name: "large file",
			setup: func() (io.Reader, Config) {
				largeContent := strings.Repeat("test content\n", 1024*1024*10) // ~100MB
				filepath, cleanup := createTestFile(t, largeContent)
				t.Cleanup(cleanup)
				file, _ := os.Open(filepath)
				return file, Config{}
			},
			wantErr: "file too large",
		},
		{
			name: "invalid UUID version",
			setup: func() (io.Reader, Config) {
				return strings.NewReader("some content"), Config{UUIDVersion: 6}
			},
			wantErr: "invalid UUID version",
		},
		{
			name: "context cancelled",
			setup: func() (io.Reader, Config) {
				// Create a large enough input to ensure processing time
				return strings.NewReader(strings.Repeat("test content\n", 1000)), Config{}
			},
			wantErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, config := tt.setup()
			ext, err := New(config)
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("New() error = %v, want error containing %q", err, tt.wantErr)
				return
			}
			if err == nil {
				ctx, cancel := context.WithCancel(context.Background())
				if tt.name == "context cancelled" {
					cancel()
				}
				defer cancel()
				_, err = ext.Extract(ctx, reader)
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Extract() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestExtractor_ExtractWithValidation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		config Config
		want   Results
	}{
		{
			name: "validate email format",
			input: `valid@example.com
invalid@.com
@invalid.com
noat.com`,
			config: Config{ExtractEmails: true},
			want: Results{
				Emails: map[string]bool{
					"valid@example.com": true,
				},
			},
		},
		{
			name: "validate domain format",
			input: `https://valid.com
https://.invalid
http://invalid.
ftp://invalid.com`,
			config: Config{ExtractDomains: true},
			want: Results{
				Domains: map[string]bool{
					"valid.com": true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := New(tt.config)
			if err != nil {
				t.Fatalf("Failed to create extractor: %v", err)
			}

			reader := strings.NewReader(tt.input)
			got, err := ext.Extract(context.Background(), reader)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkExtractor_Extract(b *testing.B) {
	// Create test data
	var builder strings.Builder
	testData := `https://example.com/users?id=123&token=abc
user@example.com
192.168.1.1
550e8400-e29b-41d4-a716-446655440000
`
	// Repeat the test data to create a larger file
	for i := 0; i < 1000; i++ {
		builder.WriteString(testData)
	}

	config := Config{
		UUIDVersion:    4,
		ExtractEmails:  true,
		ExtractDomains: true,
		ExtractIPs:     true,
		ExtractParams:  true,
	}

	ext, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create extractor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(builder.String())
		_, err := ext.Extract(context.Background(), reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				UUIDVersion: 4,
			},
			wantErr: false,
		},
		{
			name: "invalid UUID version",
			config: Config{
				UUIDVersion: 6,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ext == nil {
				t.Error("New() returned nil extractor without error")
			}
		})
	}
}
