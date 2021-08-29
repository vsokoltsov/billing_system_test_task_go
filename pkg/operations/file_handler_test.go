package operations

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type FailedFileStore struct {
}

func (ffs FailedFileStore) Create(path string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, fmt.Errorf("error file creation")
}

func TestFileHandlerSuccessCreateFile(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	defer os.Remove("report.json")

	_, err := fh.Create("json")
	if err != nil {
		t.Errorf("File was not created: %s", err)
	}
}

func TestFileHandlerFailedCreateFile(t *testing.T) {
	fh := FileHandler{
		fileStorage: FailedFileStore{},
	}
	defer os.Remove("report.json")

	_, err := fh.Create("json")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestFileHandlerSuccessCreateFileCSVFormat(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	defer os.Remove("report.csv")

	params, err := fh.Create("csv")
	if err != nil {
		t.Errorf("File was not created: %s", err)
	}

	if params.csvWriter == nil {
		t.Errorf("Expected csvWriter to be non-empty")
	}
}

func TestFileHandlerSuccessCreateMarshallerCSV(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := csv.NewWriter(f)
	marshaller := fh.CreateMarshaller(f, "csv", csvWriter)
	if reflect.TypeOf(marshaller) != reflect.TypeOf(&CSVHandler{}) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(CSVHandler{}), reflect.TypeOf(marshaller))
	}
}

func TestFileHandlerSuccessCreateMarshallerJSON(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := csv.NewWriter(f)
	marshaller := fh.CreateMarshaller(f, "json", csvWriter)
	if reflect.TypeOf(marshaller) != reflect.TypeOf(&JSONHandler{}) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(&JSONHandler{}), reflect.TypeOf(marshaller))
	}
}

func TestNewFileHandlerFunction(t *testing.T) {
	storage := FileStorage{}
	handler := NewFileHandler(storage)
	if reflect.TypeOf(handler.fileStorage) != reflect.TypeOf(storage) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(storage), reflect.TypeOf(handler.fileStorage))
	}
}
