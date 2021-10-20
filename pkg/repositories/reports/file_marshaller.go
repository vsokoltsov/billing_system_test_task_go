package reports

import (
	"billing_system_test_task/pkg/entities"
	"fmt"
	"io"
	"strconv"
	"sync"
)

// FileMarshallingManager defines contracts for file marshalling
type FileMarshallingManager interface {
	MarshallOperation(operation *entities.WalletOperation) (*MarshalledResult, error)
	WriteToFile(mr *MarshalledResult) error
}

// FileMarshallingManager represents methods for csv writing
type CSVWriter interface {
	Write(record []string) error
	Flush()
}

// JSONHandler implements FileMarshallingManager interface for json format
type JSONHandler struct {
	file     io.Writer
	mu       *sync.Mutex
	marshall func(v interface{}) ([]byte, error)
}

func NewJSONHandler(file io.Writer, mu *sync.Mutex, marshall func(v interface{}) ([]byte, error)) *JSONHandler {
	return &JSONHandler{
		file:     file,
		mu:       mu,
		marshall: marshall,
	}
}

// MarshallOperation marshal entities.WalletOperation instance to json
func (jh *JSONHandler) MarshallOperation(operation *entities.WalletOperation) (*MarshalledResult, error) {
	jsonBytes, jsonMarshallErr := jh.marshall(operation)
	if jsonMarshallErr != nil {
		return nil, fmt.Errorf("error of json marshalling: %s", jsonMarshallErr)
	}
	newLine := []byte("\n")
	data := append(jsonBytes, newLine...)
	return &MarshalledResult{
		id:   operation.ID,
		data: data,
	}, nil
}

// WriteToFile writes given marshall result to json file
func (jh *JSONHandler) WriteToFile(mr *MarshalledResult) error {
	var syncErr error
	bytesData := mr.data.([]byte)
	jh.mu.Lock()
	_, writeErr := jh.file.Write(bytesData)
	if writeErr != nil {
		return fmt.Errorf("write file error: %s", syncErr)
	}
	jh.mu.Unlock()
	return nil
}

// JSONHandler implements FileMarshallingManager interface for csv format
type CSVHandler struct {
	csvWriter CSVWriter
	mu        *sync.Mutex
}

// MarshallOperation marshal entities.WalletOperation instance to csv
func (ch *CSVHandler) MarshallOperation(operation *entities.WalletOperation) (*MarshalledResult, error) {
	idStr := strconv.Itoa(operation.ID)
	walletFromStr := strconv.Itoa(int(operation.WalletFrom.Int32))
	walletToStr := strconv.Itoa(int(operation.WalletTo))
	amountStr := operation.Amount.String()
	createdAtStr := operation.CreatedAt.String()
	row := []string{
		idStr,
		operation.Operation,
		walletFromStr,
		walletToStr,
		amountStr,
		createdAtStr,
	}
	return &MarshalledResult{
		id:   operation.ID,
		data: row,
	}, nil
}

// WriteToFile writes given marshall result to csv file
func (ch *CSVHandler) WriteToFile(mr *MarshalledResult) error {
	row := mr.data.([]string)
	ch.mu.Lock()
	csvWriteErr := ch.csvWriter.Write(row)
	if csvWriteErr != nil {
		return fmt.Errorf("error fo csv writing: %s", csvWriteErr)
	}
	ch.mu.Unlock()
	return nil
}
