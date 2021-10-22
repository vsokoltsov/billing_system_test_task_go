package reports

import (
	"billing_system_test_task/internal/entities"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

type FileWithMetadata interface {
	Read(p []byte) (n int, err error)
	Stat() (os.FileInfo, error)
}

// FileHandlingManager represents interface for file handler
type FileHandlingManager interface {
	Create(format string) (*entities.FileParams, error)
	CreateMarshaller(file *os.File, format string, csvWriter CSVWriter) (FileMarshallingManager, error)
	GetFileMetadata(file FileWithMetadata) (*entities.Metadata, error)
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

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// NewFileHandler returns new instance of FileHandler
func NewFileHandler(storage FileStorageManager) *FileHandler {
	return &FileHandler{
		fileStorage: storage,
	}
}

// Create creates new file
func (fs FileStorage) Create(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// Create file with attributes
func (fh FileHandler) Create(format string) (*entities.FileParams, error) {
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

	return &entities.FileParams{
		Name:      name,
		Path:      path,
		File:      f,
		CsvWriter: csvWriter,
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

// GetFileMetadata retrieves file's metadata
func (fh FileHandler) GetFileMetadata(file FileWithMetadata) (*entities.Metadata, error) {
	header := make([]byte, 512)
	_, readErr := file.Read(header)
	if readErr != nil {
		return nil, fmt.Errorf("error of file header's reading: %s", readErr)
	}
	stat, statErr := file.Stat()
	if statErr != nil {
		return nil, fmt.Errorf("error of file stats's receiving: %s", statErr)
	}
	size := strconv.FormatInt(stat.Size(), 10)
	contentType := http.DetectContentType(header)
	return &entities.Metadata{
		Size:        size,
		ContentType: contentType,
	}, nil
}
