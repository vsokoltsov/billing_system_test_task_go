package operations

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestJSONHandlerFileMarshallerSuccessMarshallOperation(t *testing.T) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID: 1,
	}
	handler := JSONHandler{
		file:     os.Stdout,
		mu:       mu,
		marshall: json.Marshal,
	}
	mr, _ := handler.MarshallOperation(operation)
	if mr.id != operation.ID {
		t.Errorf("ID of marshalled result does not matched. Expected: %d, got: %d", operation.ID, mr.id)
	}
}

func TestJSONHandlerFileMarshallerFailedMarshallOperation(t *testing.T) {

	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID: 1,
	}
	handler := JSONHandler{
		file: os.Stdout,
		mu:   mu,
		marshall: func(v interface{}) ([]byte, error) {
			return nil, fmt.Errorf("Marshall error")
		},
	}
	_, mrErr := handler.MarshallOperation(operation)
	if mrErr == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestJSONHandlerFileMarshallerSuccessWriteFile(t *testing.T) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID: 1,
	}
	f, _ := os.CreateTemp("", "_example_file")
	handler := JSONHandler{
		file:     f,
		mu:       mu,
		marshall: json.Marshal,
	}
	mr, _ := handler.MarshallOperation(operation)
	err := handler.WriteToFile(mr)
	if err != nil {
		t.Errorf("Expected does not receive error, but error received: %s", err)
	}
}

type ErrorFile struct{}

func (ef ErrorFile) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("Error of file writing")
}

func TestJSONHandlerFileMarshallFailedWriteFile(t *testing.T) {

	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID: 1,
	}
	handler := JSONHandler{
		file:     ErrorFile{},
		mu:       mu,
		marshall: json.Marshal,
	}
	mr, _ := handler.MarshallOperation(operation)
	err := handler.WriteToFile(mr)
	if err == nil {
		t.Errorf("Expected receive error, but error received it")
	}
}
