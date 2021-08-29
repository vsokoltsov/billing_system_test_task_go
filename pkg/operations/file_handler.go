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

type IFileHandling interface {
	Create(format string) (*FileParams, error)
	CreateMarshaller(file *os.File, format string, csvWriter *csv.Writer) IFileMarshaller
}

type IFileStorage interface {
	Create(path string, flag int, perm os.FileMode) (*os.File, error)
}

type FileHandler struct {
	fileStorage IFileStorage
}

type FileStorage struct{}

func NewFileHandler(storage IFileStorage) FileHandler {
	return FileHandler{
		fileStorage: storage,
	}
}

func (fs FileStorage) Create(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

type FileParams struct {
	f         *os.File
	path      string
	name      string
	csvWriter *csv.Writer
}

func (fh FileHandler) Create(format string) (*FileParams, error) {
	var (
		path      string
		name      string
		csvWriter *csv.Writer
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

func (fh FileHandler) CreateMarshaller(file *os.File, format string, csvWriter *csv.Writer) IFileMarshaller {
	var (
		mu          = &sync.Mutex{}
		fileHandler IFileMarshaller
	)

	if format == "csv" {
		headers := []string{
			"id", "operation", "wallet_from", "wallet_to", "amount", "created_at",
		}
		csvWriter.Write(headers)
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
	return fileHandler
}
