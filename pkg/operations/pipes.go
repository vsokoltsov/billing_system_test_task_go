package operations

import (
	"context"
	"fmt"
	"sync"
)

type ReadPipe struct {
	or     IWalletOperationRepo
	errors chan error
	wg     *sync.WaitGroup
	params *ListParams
	ctx    context.Context
}

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

type MarshallPipe struct {
	wg     *sync.WaitGroup
	fm     IFileMarshaller
	errors chan error
}

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

type WritePipe struct {
	wg     *sync.WaitGroup
	fm     IFileMarshaller
	errors chan error
}

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
