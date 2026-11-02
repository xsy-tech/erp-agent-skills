package distpack

import (
	"archive/tar"
	"archive/zip"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	FormatTarGzip = "tar.gz"
	FormatZIP     = "zip"
)

var archiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

type Options struct {
	Format     string
	SourceDir  string
	OutputPath string
	RootName   string
	BinaryName string
}

type archiveEntry struct {
	name string
	mode os.FileMode
	data []byte
}

func Write(options Options) (err error) {
	entries, err := loadEntries(options)
	if err != nil {
		return err
	}

	output, err := os.Create(options.OutputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	complete := false
	defer func() {
		if closeErr := output.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close output: %w", closeErr)
		}
		if !complete || err != nil {
			_ = os.Remove(options.OutputPath)
		}
	}()

	switch options.Format {
	case FormatTarGzip:
		err = writeTarGzip(output, entries)
	case FormatZIP:
		err = writeZIP(output, entries)
	default:
		err = fmt.Errorf("unsupported format %q", options.Format)
	}
	if err != nil {
		return err
	}
	complete = true
	return nil
}

func loadEntries(options Options) ([]archiveEntry, error) {
	if options.SourceDir == "" || options.OutputPath == "" {
		return nil, fmt.Errorf("source and output paths are required")
	}
	if !validArchiveName(options.RootName) {
		return nil, fmt.Errorf("invalid archive root %q", options.RootName)
	}
	if !validArchiveName(options.BinaryName) {
		return nil, fmt.Errorf("invalid binary name %q", options.BinaryName)
	}

	files := []struct {
		name string
		mode os.FileMode
	}{
		{name: "LICENSE", mode: 0o644},
		{name: "README.md", mode: 0o644},
		{name: "manifest.json", mode: 0o644},
		{name: options.BinaryName, mode: 0o755},
	}
	entries := make([]archiveEntry, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(options.SourceDir, file.name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file.name, err)
		}
		entries = append(entries, archiveEntry{
			name: path.Join(options.RootName, file.name),
			mode: file.mode,
			data: data,
		})
	}
	return entries, nil
}

func validArchiveName(name string) bool {
	return name != "" && name != "." && path.Clean(name) == name &&
		!strings.HasPrefix(name, "/") && !strings.Contains(name, "\\")
}

func writeTarGzip(output io.Writer, entries []archiveEntry) error {
	gzipWriter, err := gzip.NewWriterLevel(output, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	gzipWriter.Header.ModTime = time.Time{}
	gzipWriter.Header.Name = ""
	gzipWriter.Header.Comment = ""
	gzipWriter.Header.Extra = nil
	gzipWriter.Header.OS = 255

	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     int64(entry.mode),
			Size:     int64(len(entry.data)),
			ModTime:  archiveTime,
			Typeflag: tar.TypeReg,
			Uid:      0,
			Gid:      0,
			Uname:    "root",
			Gname:    "root",
			Format:   tar.FormatUSTAR,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			_ = gzipWriter.Close()
			return fmt.Errorf("write tar header %s: %w", entry.name, err)
		}
		if _, err := tarWriter.Write(entry.data); err != nil {
			_ = gzipWriter.Close()
			return fmt.Errorf("write tar payload %s: %w", entry.name, err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		_ = gzipWriter.Close()
		return fmt.Errorf("close tar writer: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("close gzip writer: %w", err)
	}
	return nil
}

func writeZIP(output io.Writer, entries []archiveEntry) error {
	zipWriter := zip.NewWriter(output)
	zipWriter.RegisterCompressor(zip.Deflate, func(writer io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(writer, flate.BestCompression)
	})
	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:   entry.name,
			Method: zip.Deflate,
		}
		header.SetMode(entry.mode)
		header.SetModTime(archiveTime)
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("write zip header %s: %w", entry.name, err)
		}
		if _, err := writer.Write(entry.data); err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("write zip payload %s: %w", entry.name, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}
	return nil
}
