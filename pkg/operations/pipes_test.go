package operations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

// Test success read pipe rows receiving
func TestSuccessReadPipe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	errCh := make(chan error)
	or := NewWalletOperationRepo(db)
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(op.ID, op.Operation, op.WalletFrom, op.WalletTo, op.Amount, op.CreatedAt)
	mock.ExpectQuery("select").WillReturnRows(rows)

	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: nil,
		ctx:    ctx,
		errors: errCh,
	}

	wg.Add(1)
	go readPipe.Call(in, out)
	wg.Wait()

	res := <-out
	chanOp := res.(*WalletOperation)
	if chanOp.ID != op.ID {
		t.Errorf("Received ID do not match. Expected %d, got %d", op.ID, chanOp.ID)
	}
}

// Test failed read pipe rows receiving (rows error)
func TestFailedReadPipeRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	errCh := make(chan error, 1)
	or := NewWalletOperationRepo(db)
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)

	mock.ExpectQuery("select").WillReturnError(fmt.Errorf("log error"))

	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: nil,
		ctx:    ctx,
		errors: errCh,
	}

	wg.Add(1)
	go readPipe.Call(in, out)
	wg.Wait()
	// close(out)

	err = <-errCh
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	nilData := <-out
	if nilData != nil {
		t.Errorf("Expected nil value, got %s", nilData)
	}
}

// Test failed read pipe rows receiving (scan error)
func TestFailedReadPipeScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	errCh := make(chan error, 1)
	or := NewWalletOperationRepo(db)
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)

	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(nil, op.Operation, op.WalletFrom, op.WalletTo, op.Amount, op.CreatedAt).RowError(1, fmt.Errorf("Scan error"))
	mock.ExpectQuery("select").WillReturnRows(rows)

	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: nil,
		ctx:    ctx,
		errors: errCh,
	}

	wg.Add(1)
	go readPipe.Call(in, out)
	wg.Wait()

	err = <-errCh
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	nilData := <-out
	if nilData != nil {
		t.Errorf("Expected nil value, got %s", nilData)
	}
}

// Test failed read pipe rows receiving (query returns emtpy result)
func TestFailedReadPipeEmptyQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	errCh := make(chan error, 1)
	or := NewWalletOperationRepo(db)
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	mock.ExpectQuery("select").WillReturnRows(rows)

	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: nil,
		ctx:    ctx,
		errors: errCh,
	}

	wg.Add(1)
	go readPipe.Call(in, out)
	wg.Wait()

	nilRes := <-out
	if nilRes != nil {
		t.Errorf("Expected nil result, got %s", nilRes)
	}
}

// Test success marshall pipe json marshalling
func TestSuccessMarshallPipe(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	marshallPipe := MarshallPipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(&MarshalledResult{
		id:   op.ID,
		data: op,
	}, nil)

	in <- &op
	wg.Add(1)
	go marshallPipe.Call(in, out)
	close(in)
	wg.Wait()

	res := <-out
	chanOp := res.(*MarshalledResult)
	if chanOp.id != op.ID {
		t.Errorf("Received ID do not match. Expected %d, got %d", op.ID, chanOp.id)
	}
}

// Test failed marshall pipe json marshalling (error marshalling)
func TestFailedMarshallPipeErrorMarshalling(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	marshallPipe := MarshallPipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(nil, fmt.Errorf("marshall error"))

	in <- &op
	wg.Add(1)
	go marshallPipe.Call(in, out)
	wg.Wait()

	err := <-errCh
	res := <-out
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if res != nil {
		t.Errorf("Expected nil result, got %s", res)
	}
}

// Test failed marshall pipe json marshalling (previous step cancelled)
func TestFailedMarshallPipePrevPipeCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)

	marshallPipe := MarshallPipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	in <- nil

	wg.Add(1)
	go marshallPipe.Call(in, out)
	wg.Wait()

	res := <-out

	if res != nil {
		t.Errorf("Expected nil result, got %s", res)
	}
}

// Test success write pipe file writing
func TestSuccessWritePipe(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	mr := MarshalledResult{
		id:   op.ID,
		data: op,
	}

	writePipe := WritePipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	in <- &mr
	mockFileMarshaller.EXPECT().WriteToFile(&mr).Return(nil)

	wg.Add(1)
	go writePipe.Call(in, out)
	close(in)
	wg.Wait()

	select {
	case err := <-errCh:
		t.Errorf("Unexpected error %s", err)
	case nilData := <-out:
		t.Errorf("Unexpected nil data from out %s", nilData)
	default:
		fmt.Println("Success file write")
	}
}

// Test failed write pipe file writing (write file error)
func TestFailedWritePipe(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	mr := MarshalledResult{
		id:   op.ID,
		data: op,
	}

	writePipe := WritePipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	in <- &mr
	mockFileMarshaller.EXPECT().WriteToFile(&mr).Return(fmt.Errorf("File error"))

	wg.Add(1)
	go writePipe.Call(in, out)
	close(in)
	wg.Wait()

	err := <-errCh
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSuccessPipelineRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	or := NewWalletOperationRepo(db)
	ctx := context.Background()
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1, Valid: true},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	mr := MarshalledResult{
		id:   op.ID,
		data: op,
	}

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(op.ID, op.Operation, op.WalletFrom.Int32, op.WalletTo, op.Amount, op.CreatedAt)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(&mr, nil)
	mockFileMarshaller.EXPECT().WriteToFile(&mr).Return(nil)

	oProcessor := OperationsProcessesManager{}
	processErr := oProcessor.Process(ctx, or, nil, mockFileMarshaller)
	if processErr != nil {
		t.Errorf("Unexpected error: %s", processErr)
	}
}

func TestFailedPipelineRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	or := NewWalletOperationRepo(db)
	ctx := context.Background()
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1, Valid: true},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(op.ID, op.Operation, op.WalletFrom.Int32, op.WalletTo, op.Amount, op.CreatedAt)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(nil, fmt.Errorf("marshall error"))

	oProcessor := OperationsProcessesManager{}
	processErr := oProcessor.Process(ctx, or, nil, mockFileMarshaller)
	if processErr == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(processErr.Error(), "marshall error") {
		t.Errorf("Wrong error message string")
	}
}

func BenchmarkPipeline(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	or := NewWalletOperationRepo(db)
	ctx := context.Background()
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1, Valid: true},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}
	mr := MarshalledResult{
		id:   op.ID,
		data: op,
	}

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(op.ID, op.Operation, op.WalletFrom.Int32, op.WalletTo, op.Amount, op.CreatedAt)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(&mr, nil)
	mockFileMarshaller.EXPECT().WriteToFile(&mr).Return(nil)

	oProcessor := OperationsProcessesManager{}

	for i := 0; i < b.N; i++ {
		_ = oProcessor.Process(ctx, or, nil, mockFileMarshaller)
	}
}

func BenchmarkReadPipe(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	errCh := make(chan error)
	or := NewWalletOperationRepo(db)
	wg := &sync.WaitGroup{}
	ctx := context.Background()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
	rows = rows.AddRow(op.ID, op.Operation, op.WalletFrom, op.WalletTo, op.Amount, op.CreatedAt)
	mock.ExpectQuery("select").WillReturnRows(rows)

	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: nil,
		ctx:    ctx,
		errors: errCh,
	}

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go readPipe.Call(in, out)
	}
}

func BenchmarkMarshallPipe(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	marshallPipe := MarshallPipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(&MarshalledResult{
		id:   op.ID,
		data: op,
	}, nil).AnyTimes()

	for i := 0; i < b.N; i++ {
		in <- &op
		wg.Add(1)
		go marshallPipe.Call(in, out)
	}
}

func BenchmarkWritePipe(b *testing.B) {
	ctrl := gomock.NewController(b)
	mockFileMarshaller := NewMockFileMarshallingManager(ctrl)
	errCh := make(chan error, 1)
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	op := WalletOperation{
		ID:         1,
		Operation:  "deposit",
		WalletFrom: sql.NullInt32{Int32: 1},
		WalletTo:   2,
		Amount:     decimal.NewFromInt(100),
		CreatedAt:  time.Now(),
	}

	marshallPipe := MarshallPipe{
		wg:     wg,
		errors: errCh,
		fm:     mockFileMarshaller,
	}

	mockFileMarshaller.EXPECT().MarshallOperation(&op).Return(&MarshalledResult{
		id:   op.ID,
		data: op,
	}, nil).AnyTimes()

	for i := 0; i < b.N; i++ {
		in <- &op
		wg.Add(1)
		go marshallPipe.Call(in, out)
	}
}
