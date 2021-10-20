package entities

import "os"

// FileParams represents information about created file
type FileParams struct {
	File      *os.File
	Path      string
	Name      string
	CsvWriter CSVWriter
}

type Metadata struct {
	Size        string
	ContentType string
}

type CSVWriter interface {
	Write(record []string) error
	Flush()
}

type FileMetadata struct {
	File        *os.File
	Path        string
	Size        string
	ContentType string
}
