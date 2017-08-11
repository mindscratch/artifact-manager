package core

import (
	"fmt"
	"io"
	"os"

	"github.com/mholt/archiver"
	filetype "gopkg.in/h2non/filetype.v1"
)

// SaveFile writes the contents from reader to a new file named fileName.
//
// The expectedLength is used to verify the entire contents were successfully written.
func SaveFile(fileName string, reader io.Reader, expectedLength int64) error {
	// create a file to copy the request contents into
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("unable to create file to write content into: %v", err)
	}
	defer f.Close()

	// copy the request body into the file
	written, err := io.Copy(f, reader)
	if err != nil {
		return fmt.Errorf("failed to write content to %s: %v", f.Name(), err)
	}
	if written != expectedLength {
		return fmt.Errorf("failed to write entire content to %s, wrote=%d bytes, expected=%d bytes", f.Name(), written, expectedLength)
	}

	return nil
}

// ExtractFile file into a directory provided by the extractIntoDir argument.
func ExtractFile(file, extractIntoDir string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if isArchive(f) {
		err = extract(file, extractIntoDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// Symlink creates a symlink named dst pointing to src.
func Symlink(src, dst string) error {
	return os.Symlink(src, dst)
}

func isArchive(reader io.Reader) bool {
	buf := make([]byte, 512)
	_, err := reader.Read(buf)
	if err == nil {
		return filetype.IsArchive(buf)
	}
	return false
}

func extract(fileName, outputDir string) error {
	for _, archiverImpl := range archiver.SupportedFormats {
		if archiverImpl.Match(fileName) {
			return archiverImpl.Open(fileName, outputDir)
		}
	}
	return fmt.Errorf("unsupported file type: %s", fileName)
}
