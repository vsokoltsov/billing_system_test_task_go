package operations

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

type IFileMarshaller interface {
	MarshallOperation(operation *WalletOperation) (*MarshalledResult, error)
	WriteToFile(mr *MarshalledResult) error
}

type CSVWriter interface {
	Write(record []string) error
	Flush()
}

type JSONHandler struct {
	file     io.Writer
	mu       *sync.Mutex
	marshall func(v interface{}) ([]byte, error)
}

func (jh *JSONHandler) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
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

type CSVHandler struct {
	csvWriter CSVWriter
	mu        *sync.Mutex
}

func (ch *CSVHandler) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
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
