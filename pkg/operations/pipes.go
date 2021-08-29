package operations

import (
	"billing_system_test_task/pkg/pipeline"
	"context"
	"fmt"
	"sync"
)

// IOperationsProcessor defines operations for processing WalletOperation
type IOperationsProcessor interface {
	Process(ctx context.Context, or IWalletOperationRepo, listParams *ListParams, marshaller IFileMarshaller) error
}

// OperationsProcessor represents IOperationsProcessor interface
type OperationsProcessor struct{}

func (op OperationsProcessor) Process(ctx context.Context, or IWalletOperationRepo, listParams *ListParams, marshaller IFileMarshaller) error {
	var (
		wg     = &sync.WaitGroup{}
		errors = make(chan error, 1)
	)
	readPipe := ReadPipe{
		or:     or,
		wg:     wg,
		params: listParams,
		ctx:    ctx,
		errors: errors,
	}
	marshallPipe := MarshallPipe{
		wg:     wg,
		fm:     marshaller,
		errors: errors,
	}
	writePipe := WritePipe{
		wg:     wg,
		fm:     marshaller,
		errors: errors,
	}
	pipes := []pipeline.Pipe{
		readPipe,
		marshallPipe,
		writePipe,
	}

	wg.Add(len(pipes))
	pipeline.ExecutePipeline(pipes...)
	wg.Wait()

	select {
	case err := <-errors:
		return fmt.Errorf("operations read failed: %s", err)
	default:
		return nil
	}
}

// ReadPipe represents reading part of pipeline
type ReadPipe struct {
	or     IWalletOperationRepo
	errors chan error
	wg     *sync.WaitGroup
	params *ListParams
	ctx    context.Context
}

// Call reads rows from database and pass them further throught the pipeline
func (rp ReadPipe) Call(in, out chan interface{}) {
	defer rp.wg.Done()
	var counter int

	rows, rowsErr := rp.or.List(rp.ctx, rp.params)
	if rowsErr != nil {
		rp.errors <- fmt.Errorf("error of row retrieving: %s", rowsErr)
		out <- nil
		return
	}
	defer rows.Close()

	for rows.Next() {
		operation := WalletOperation{}
		scanErr := rows.Scan(&operation.ID, &operation.Operation, &operation.WalletFrom, &operation.WalletTo, &operation.Amount, &operation.CreatedAt)
		if scanErr != nil {
			rp.errors <- fmt.Errorf("error of row scanning: %s", scanErr)
			out <- nil
			return
		}
		counter++
		out <- &operation
	}
	if counter == 0 {
		out <- nil
	}
}

// MarshallPipe represents marshalling part of pipeline (to csv or json)
type MarshallPipe struct {
	wg     *sync.WaitGroup
	fm     IFileMarshaller
	errors chan error
}

// Call marshall received rows to csv or json
func (mp MarshallPipe) Call(in, out chan interface{}) {
	defer mp.wg.Done()

	for chunk := range in {
		if chunk == nil {
			out <- nil
			return
		}
		operation := chunk.(*WalletOperation)
		mr, mrErr := mp.fm.MarshallOperation(operation)
		if mrErr != nil {
			out <- nil
			err := fmt.Errorf("[ERROR] Marshalling error: %s", mrErr)
			mp.errors <- err
			return
		}
		out <- mr
	}
}

// WritePipe represents writing to file part of pipeline
type WritePipe struct {
	wg     *sync.WaitGroup
	fm     IFileMarshaller
	errors chan error
}

// Call write receive marshalled items to file
func (wp WritePipe) Call(in, out chan interface{}) {
	defer wp.wg.Done()

	for chunk := range in {
		if chunk != nil {
			mr := chunk.(*MarshalledResult)
			writeErr := wp.fm.WriteToFile(mr)
			if writeErr != nil {
				wp.errors <- fmt.Errorf("[ERROR] Write to file error: %s", writeErr)
				return
			}
		}
	}
}

type MarshalledResult struct {
	id   int
	data interface{}
}
