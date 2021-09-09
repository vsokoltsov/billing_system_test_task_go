package operations

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
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

type ErrorCSVFile struct{}

func (ef ErrorFile) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("Error of file writing")
}

func (ecf ErrorCSVFile) Write(record []string) error {
	return fmt.Errorf("Error of file writing")
}

func (ecf ErrorCSVFile) Flush() {}

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

func TestCSVHandlerFileMarshallSuccessMarshallOperation(t *testing.T) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	csvWriter := csv.NewWriter(os.Stdout)
	handler := CSVHandler{
		mu:        mu,
		csvWriter: csvWriter,
	}
	mr, mrErr := handler.MarshallOperation(operation)
	if mrErr != nil {
		t.Errorf("Unexpected error: %s", mrErr)
	}
	if mr.id != operation.ID {
		t.Errorf("ID of marshalled result does not matched. Expected: %d, got: %d", operation.ID, mr.id)
	}
}

func TestCSVHandlerFileMarshallSuccessWriteFile(t *testing.T) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	csvWriter := ErrorCSVFile{}
	handler := CSVHandler{
		mu:        mu,
		csvWriter: csvWriter,
	}
	mr, _ := handler.MarshallOperation(operation)
	writeErr := handler.WriteToFile(mr)
	if writeErr == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestCSVHandlerFileMarshallSuccessWroteFile(t *testing.T) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	csvWriter := csv.NewWriter(os.Stdout)
	handler := CSVHandler{
		mu:        mu,
		csvWriter: csvWriter,
	}
	mr, _ := handler.MarshallOperation(operation)
	writeErr := handler.WriteToFile(mr)
	if writeErr != nil {
		t.Errorf("Expected nil, got error: %s", writeErr)
	}
}

func BenchmarkMarshallOperationJSON(b *testing.B) {
	wo := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	f, _ := os.CreateTemp("", "_example_file")
	handler := JSONHandler{
		file:     f,
		mu:       &sync.Mutex{},
		marshall: json.Marshal,
	}

	for i := 0; i < b.N; i++ {
		handler.MarshallOperation(wo)
	}
}

func BenchmarkMarshallOperationCSV(b *testing.B) {
	mu := &sync.Mutex{}
	operation := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	csvWriter := csv.NewWriter(os.Stdout)
	handler := CSVHandler{
		mu:        mu,
		csvWriter: csvWriter,
	}

	for i := 0; i < b.N; i++ {
		handler.MarshallOperation(operation)
	}
}

func BenchmarkWriteToFileJSON(b *testing.B) {
	wo := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	f, _ := os.CreateTemp("", "_example_file")
	handler := JSONHandler{
		file:     f,
		mu:       &sync.Mutex{},
		marshall: json.Marshal,
	}
	mr, _ := handler.MarshallOperation(wo)
	for i := 0; i < b.N; i++ {
		handler.WriteToFile(mr)
	}
}

func BenchmarkWriteToFileCSV(b *testing.B) {
	mu := &sync.Mutex{}
	wo := &WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   1,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	f, _ := os.CreateTemp("", "_example_file")
	csvWriter := csv.NewWriter(f)
	handler := CSVHandler{
		mu:        mu,
		csvWriter: csvWriter,
	}
	mr, _ := handler.MarshallOperation(wo)
	for i := 0; i < b.N; i++ {
		handler.WriteToFile(mr)
	}
}
