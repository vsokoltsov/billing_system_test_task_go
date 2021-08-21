package operations

import (
	"context"
	"log"
	"sync"
)

type ReadPipe struct {
	or     IWalletOperationRepo
	oc     chan *WalletOperation
	wg     *sync.WaitGroup
	params *ListParams
	ctx    context.Context
}

func (rp ReadPipe) Call(in, out chan interface{}) {
	defer rp.wg.Done()

	rows, rowsErr := rp.or.List(rp.ctx, rp.params)
	if rowsErr != nil {
		log.Printf("Error of row retrieving: %s", rowsErr)
		return
	}
	for rows.Next() {
		operation := WalletOperation{}
		scanErr := rows.Scan(&operation.ID, &operation.Operation, &operation.WalletFrom, &operation.WalletTo, &operation.Amount, &operation.CreatedAt)
		if scanErr != nil {
			log.Fatalf("Error of row scanning: %s", scanErr)
		}
		out <- &operation
	}
}

type MarshallPipe struct {
	wg *sync.WaitGroup
	fm IFileMarshaller
}

func (mp MarshallPipe) Call(in, out chan interface{}) {
	defer mp.wg.Done()

	for chunk := range in {
		operation := chunk.(*WalletOperation)
		mr, mrErr := mp.fm.MarshallOperation(operation)
		if mrErr != nil {
			log.Printf("[ERROR] Marshalling error: %s", mrErr)
		} else {
			out <- *mr
		}
	}
}

type WritePipe struct {
	wg *sync.WaitGroup
	fm IFileMarshaller
}

func (wp WritePipe) Call(in, out chan interface{}) {
	defer wp.wg.Done()

	for chunk := range in {
		mr := chunk.(MarshalledResult)
		wp.fm.WriteToFile(&mr)
	}
}

type MarshalledResult struct {
	id   int
	data interface{}
}
