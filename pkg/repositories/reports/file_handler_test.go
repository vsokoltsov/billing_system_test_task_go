package reports

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
)

// FailedFileStore represents implementation of store interface with error methods
type FailedFileStore struct {
}

func (ffs FailedFileStore) Create(path string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, fmt.Errorf("error file creation")
}

// Test success file creation (json format)
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

// Test failed file creation (file store error)
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

// Test success file creation (csv format)
func TestFileHandlerSuccessCreateFileCSVFormat(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	defer os.Remove("report.csv")

	params, err := fh.Create("csv")
	if err != nil {
		t.Errorf("File was not created: %s", err)
	}

	if params.CsvWriter == nil {
		t.Errorf("Expected csvWriter to be non-empty")
	}
}

// Test success operation marshalling (csv format)
func TestFileHandlerSuccessCreateMarshallerCSV(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := csv.NewWriter(f)
	marshaller, _ := fh.CreateMarshaller(f, "csv", csvWriter)
	if reflect.TypeOf(marshaller) != reflect.TypeOf(&CSVHandler{}) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(CSVHandler{}), reflect.TypeOf(marshaller))
	}
}

// Test failed operation marshalling (csv format)
func TestFileHandlerFailedCreateMarshallerCSV(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := ErrorCSVFile{}
	_, marshallErr := fh.CreateMarshaller(f, "csv", csvWriter)
	if marshallErr == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(marshallErr.Error(), "Error of file writing") {
		t.Errorf("Expected error message is not present")
	}
}

// Test success operation marshalling (json format)
func TestFileHandlerSuccessCreateMarshallerJSON(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := csv.NewWriter(f)
	marshaller, _ := fh.CreateMarshaller(f, "json", csvWriter)
	if reflect.TypeOf(marshaller) != reflect.TypeOf(&JSONHandler{}) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(&JSONHandler{}), reflect.TypeOf(marshaller))
	}
}

// Test file handling constructor
func TestNewFileHandlerFunction(t *testing.T) {
	storage := FileStorage{}
	handler := NewFileHandler(storage)
	if reflect.TypeOf(handler.fileStorage) != reflect.TypeOf(storage) {
		t.Errorf("Types mismatch. Expected: %s. Got: %s", reflect.TypeOf(storage), reflect.TypeOf(handler.fileStorage))
	}
}

// Test success receiving of file's metadata
func TestSuccessFileHandlerGetFileMetadata(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	tmpFile, _ := ioutil.TempFile(os.TempDir(), "test")
	data := []byte("Test file\n")
	_, _ = tmpFile.Write(data)
	_, _ = tmpFile.Seek(0, 0)

	res, err := fh.GetFileMetadata(tmpFile)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if res.Size != "10" {
		t.Errorf("Unexpected file size - expected 10")
	}
}

// Test failed receiving file's metadata (Read() error)
func TestFailedFileHandlerGetFileMetadataErrorRead(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}
	tmpFile, _ := ioutil.TempFile(os.TempDir(), "test")
	data := []byte("Test file\n")
	_, _ = tmpFile.Write(data)

	_, err := fh.GetFileMetadata(tmpFile)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// Test failed receiving file's metadata (Stat() error)
func TestFailedFileHandlerGetFileMetadataErrorStats(t *testing.T) {
	fh := FileHandler{
		fileStorage: FileStorage{},
	}

	ctrl := gomock.NewController(t)
	mockFM := NewMockFileWithMetadata(ctrl)
	mockFM.EXPECT().Read(gomock.Any()).Return(0, nil)
	mockFM.EXPECT().Stat().Return(nil, fmt.Errorf("file stat get error"))

	_, err := fh.GetFileMetadata(mockFM)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "file stat get error") {
		t.Errorf("Error should contain 'file stat get error' message")
	}
}
