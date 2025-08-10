package compress

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

// Compression represents the available compression types
type Compression string

const (
	CompNone Compression = "none"
	CompGzip Compression = "gzip"
	CompZstd Compression = "zstd"
)

// String implements the Stringer interface
func (c Compression) String() string {
	return string(c)
}

// IsValid checks if the compression type is valid
func (c Compression) IsValid() bool {
	switch c {
	case CompNone, CompGzip, CompZstd:
		return true
	default:
		return false
	}
}

// WriterCloser combines io.Writer and io.Closer interfaces
type WriterCloser interface {
	io.Writer
	io.Closer
}

// ReaderCloser combines io.Reader and io.Closer interfaces
type ReaderCloser interface {
	io.Reader
	io.Closer
}

// CompressedWriter wraps a writer with compression and checksum calculation
type CompressedWriter struct {
	file       *os.File
	compressor WriterCloser
	hasher     hash.Hash
	writer     io.Writer
}

// NewCompressedWriter creates a new compressed writer
func NewCompressedWriter(path string, comp Compression) (*CompressedWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(file, hasher)

	var compressor WriterCloser
	var writer io.Writer = multiWriter

	switch comp {
	case CompZstd:
		zw, err := zstd.NewWriter(multiWriter)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create zstd writer: %w", err)
		}
		compressor = zw
		writer = zw

	case CompGzip:
		gw := gzip.NewWriter(multiWriter)
		compressor = gw
		writer = gw

	case CompNone:
		compressor = &nopCloser{Writer: multiWriter}
	}

	return &CompressedWriter{
		file:       file,
		compressor: compressor,
		hasher:     hasher,
		writer:     writer,
	}, nil
}

// Write writes data through the compression chain
func (cw *CompressedWriter) Write(p []byte) (int, error) {
	return cw.writer.Write(p)
}

// Close closes all writers and returns the checksum
func (cw *CompressedWriter) Close() (string, error) {
	var err error

	// Close compressor first
	if cw.compressor != nil {
		err = cw.compressor.Close()
	}

	// Sync and close file
	if syncErr := cw.file.Sync(); syncErr != nil && err == nil {
		err = syncErr
	}

	if closeErr := cw.file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", cw.hasher.Sum(nil))

	return checksum, err
}

// CompressedReader wraps a reader with decompression
type CompressedReader struct {
	file         *os.File
	decompressor ReaderCloser
	reader       io.Reader
}

// NewCompressedReader creates a new compressed reader
func NewCompressedReader(path string, comp Compression) (*CompressedReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	var decompressor ReaderCloser
	var reader io.Reader = file

	switch comp {
	case CompZstd:
		zr, err := zstd.NewReader(file)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		decompressor = &zstdReaderCloser{zr}
		reader = zr

	case CompGzip:
		gr, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		decompressor = gr
		reader = gr

	case CompNone:
		decompressor = &nopCloser{Reader: file}
	}

	return &CompressedReader{
		file:         file,
		decompressor: decompressor,
		reader:       reader,
	}, nil
}

// Read reads data through the decompression chain
func (cr *CompressedReader) Read(p []byte) (int, error) {
	return cr.reader.Read(p)
}

// Close closes all readers
func (cr *CompressedReader) Close() error {
	var err error

	// Close decompressor first
	if cr.decompressor != nil {
		err = cr.decompressor.Close()
	}

	// Close file
	if closeErr := cr.file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

// CalculateSHA256 calculates the SHA256 checksum of a file
func CalculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// nopCloser wraps an io.Writer or io.Reader to add a no-op Close method
type nopCloser struct {
	io.Writer
	io.Reader
}

func (nc *nopCloser) Close() error {
	return nil
}

// zstdReaderCloser wraps zstd.Decoder to provide proper Close method
type zstdReaderCloser struct {
	*zstd.Decoder
}

func (zrc *zstdReaderCloser) Close() error {
	zrc.Decoder.Close()
	return nil
}
