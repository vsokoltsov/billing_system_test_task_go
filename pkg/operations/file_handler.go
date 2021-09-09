package operations

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// FileHandlingManager represents interface for file handler
type FileHandlingManager interface {
	Create(format string) (*FileParams, error)
	CreateMarshaller(file *os.File, format string, csvWriter CSVWriter) (FileMarshallingManager, error)
}

// FileStorageManager represents interface for file storage
type FileStorageManager interface {
	Create(path string, flag int, perm os.FileMode) (*os.File, error)
}

// FileHandler implements FileHandlingManager interface
type FileHandler struct {
	fileStorage FileStorageManager
}

// FileStorage implements FileStorageManager interface
type FileStorage struct{}

// NewFileHandler returns new instance of FileHandler
func NewFileHandler(storage FileStorageManager) FileHandler {
	return FileHandler{
		fileStorage: storage,
	}
}

// Create creates new file
func (fs FileStorage) Create(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// FileParams represents information about created file
type FileParams struct {
	f         *os.File
	path      string
	name      string
	csvWriter CSVWriter
}

// Create file with attributes
func (fh FileHandler) Create(format string) (*FileParams, error) {
	var (
		path      string
		name      string
		csvWriter CSVWriter
	)

	name = "report." + format
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	path = filepath.Join(basepath, name)
	f, fileOpenErr := fh.fileStorage.Create(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if fileOpenErr != nil {
		return nil, fmt.Errorf("error of creating file: %s", fileOpenErr)
	}

	if format == "csv" {
		csvWriter = csv.NewWriter(f)
	}

	return &FileParams{
		name:      name,
		path:      path,
		f:         f,
		csvWriter: csvWriter,
	}, nil
}

// CreateMarshaller returns file marshaller for particular format
func (fh FileHandler) CreateMarshaller(file *os.File, format string, csvWriter CSVWriter) (FileMarshallingManager, error) {
	var (
		mu          = &sync.Mutex{}
		fileHandler FileMarshallingManager
	)

	if format == "csv" {
		headers := []string{
			"id", "operation", "wallet_from", "wallet_to", "amount", "created_at",
		}
		writeErr := csvWriter.Write(headers)
		if writeErr != nil {
			return nil, fmt.Errorf("error fo csv writing: %s", writeErr)
		}
		fileHandler = &CSVHandler{
			csvWriter: csvWriter,
			mu:        mu,
		}
	} else if format == "json" {
		fileHandler = &JSONHandler{
			file:     file,
			mu:       mu,
			marshall: json.Marshal,
		}
	}
	return fileHandler, nil
}
