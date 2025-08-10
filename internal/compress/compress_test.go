package compress

import (
	"io"
	"os"
	"testing"
)

func TestCompressionTypes(t *testing.T) {
	tests := []struct {
		comp  Compression
		valid bool
	}{
		{CompZstd, true},
		{CompGzip, true},
		{CompNone, true},
		{"invalid", false},
	}

	for _, test := range tests {
		if test.comp.IsValid() != test.valid {
			t.Errorf("Compression %s: expected valid=%v, got %v", test.comp, test.valid, test.comp.IsValid())
		}
	}
}

func TestCompressedWriterRoundtrip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "nizam-test-*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	testData := []byte("Hello, World! This is a test of compressed data.")

	// Test each compression type
	compressions := []Compression{CompNone, CompGzip, CompZstd}
	for _, comp := range compressions {
		t.Run(string(comp), func(t *testing.T) {
			// Write compressed data
			writer, err := NewCompressedWriter(tmpFile.Name(), comp)
			if err != nil {
				t.Fatalf("Failed to create compressed writer: %v", err)
			}

			n, err := writer.Write(testData)
			if err != nil {
				t.Fatalf("Failed to write data: %v", err)
			}
			if n != len(testData) {
				t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
			}

			checksum, err := writer.Close()
			if err != nil {
				t.Fatalf("Failed to close writer: %v", err)
			}

			if checksum == "" {
				t.Error("Expected non-empty checksum")
			}

			// Read compressed data
			reader, err := NewCompressedReader(tmpFile.Name(), comp)
			if err != nil {
				t.Fatalf("Failed to create compressed reader: %v", err)
			}
			defer reader.Close()

			// Read all data using io.ReadAll for proper handling of compressed streams
			readData, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Failed to read data: %v", err)
			}
			if len(readData) != len(testData) {
				t.Errorf("Expected to read %d bytes, read %d", len(testData), len(readData))
			}

			if string(readData) != string(testData) {
				t.Errorf("Data mismatch: expected %s, got %s", string(testData), string(readData))
			}
		})
	}
}
