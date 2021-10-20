package reports

import (
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/pipeline"
	"billing_system_test_task/pkg/repositories"
	"context"
	"fmt"
	"sync"
)

// PipelineManager defines operations for processing entities.WalletOperation
type PipelineManager interface {
	Process(ctx context.Context, or repositories.OperationsManager, listParams *repositories.ListParams, marshaller FileMarshallingManager) error
}

// OperationsProcessesManager represents PipelineManager interface
type OperationsProcessesManager struct{}

// Process runs pipeline through all the stages
func (op OperationsProcessesManager) Process(ctx context.Context, or repositories.OperationsManager, listParams *repositories.ListParams, marshaller FileMarshallingManager) error {
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
	or     repositories.OperationsManager
	errors chan error
	wg     *sync.WaitGroup
	params *repositories.ListParams
	ctx    context.Context
}

// Call reads rows from database and pass them further throught the pipeline
func (rp ReadPipe) Call(in, out chan interface{}) {
	defer rp.wg.Done()
	var counter int

	rowsCh, rowsErr := rp.or.List(rp.ctx, rp.params)
	if rowsErr != nil {
		rp.errors <- fmt.Errorf("error of row retrieving: %s", rowsErr)
		out <- nil
		return
	}
	for operation := range rowsCh {
		counter++
		out <- operation
	}
	if counter == 0 {
		out <- nil
	}
}

// MarshallPipe represents marshalling part of pipeline (to csv or json)
type MarshallPipe struct {
	wg     *sync.WaitGroup
	fm     FileMarshallingManager
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
		operation := chunk.(*entities.WalletOperation)
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
	fm     FileMarshallingManager
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

// MarshalledResult represents result of marshalling operation
type MarshalledResult struct {
	id   int
	data interface{}
}
