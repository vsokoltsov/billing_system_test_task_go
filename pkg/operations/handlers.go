package operations

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

type ReadPipe struct {
	OperationsRepo IWalletOperationRepo
	jobs           <-chan ReadJob
}

func (rp *ReadPipe) Call(in, out interface{}) {

}

type MarshallPipe struct {
}

func (mp *MarshallPipe) Call(in, out interface{}) {

}

type WritePipe struct {
}

func (wp *WritePipe) Call(in, out interface{}) {

}

type ReadJob struct {
	format string
	wo     *WalletOperation
}

type MarshalledResult struct {
	id   int
	data interface{}
}

type OperationsHandler struct {
	OperationsRepo IWalletOperationRepo
}

// Create godoc
// @Summary Wallet operations
// @Description Get wallet operations logs
// @Tags operations
// @Accept  json
// @Produce application/octet-stream
// @Router /api/operations/ [get]
// @Header 200 {string} Content-Type "application/octet-stream"
// @Header 200 {string} Expires "0"
func (oh *OperationsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var (
		format         string
		csvWriter      *csv.Writer
		headers        []string
		f              *os.File
		fileOpenErr    error
		listParams     *ListParams
		operationsChan = make(chan *WalletOperation, 1)
		marshalledChan = make(chan MarshalledResult, 1)
		waitGroup      = &sync.WaitGroup{}
		errors         = make(chan error)
		fileMutex      = &sync.Mutex{}
		resultChan     = make(chan bool)
		fileHandler    IFileMarshaller
	)
	v := r.URL.Query()

	format = v.Get("format")
	pageStr := v.Get("page")
	perPageStr := v.Get("per_page")

	if format == "" {
		format = "json"
	}

	if pageStr != "" && perPageStr != "" {
		page, pageConvError := strconv.Atoi(pageStr)
		if pageConvError != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		perPage, perPageConvError := strconv.Atoi(perPageStr)
		if perPageConvError != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		listParams = &ListParams{
			page:    page,
			perPage: perPage,
		}
	}

	fileName := "report." + format
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	fullPath := filepath.Join(basepath, fileName)
	f, fileOpenErr = os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if fileOpenErr != nil {
		fmt.Println(fileOpenErr)
		return
	}
	defer func(filePathInfo string, fileData *os.File) {
		fileData.Close()
		os.Remove(filePathInfo)
	}(fullPath, f)

	if format == "csv" {
		csvWriter = csv.NewWriter(f)

		headers = []string{
			"id", "operation", "wallet_from", "wallet_to", "amount", "created_at",
		}
		csvWriter.Write(headers)
		fileHandler = &CSVHandler{
			csvWriter: csvWriter,
			mu:        fileMutex,
		}
	} else if format == "json" {
		fileHandler = &JSONHandler{
			file: f,
			mu:   fileMutex,
		}
	}
	waitGroup.Add(1)
	go func(handler *OperationsHandler, oc chan *WalletOperation, wg *sync.WaitGroup, errors chan error) {
		rows, rowsErr := oh.OperationsRepo.List(ctx, listParams)

		defer close(oc)
		defer wg.Done()

		if rowsErr != nil {
			fmt.Println(rowsErr)
			return
		}
		for rows.Next() {
			operation := WalletOperation{}
			scanErr := rows.Scan(&operation.ID, &operation.Operation, &operation.WalletFrom, &operation.WalletTo, &operation.Amount, &operation.CreatedAt)
			if scanErr != nil {
				log.Fatalf("Error of row scanning: %s", scanErr)
			}
			oc <- &operation
		}

	}(oh, operationsChan, waitGroup, errors)

	waitGroup.Add(1)
	go func(oc chan *WalletOperation, mc chan MarshalledResult, wg *sync.WaitGroup, format string, fileHandler IFileMarshaller, errors chan error) {
		defer close(mc)
		defer wg.Done()

		for operation := range operationsChan {
			mr, mrErr := fileHandler.MarshallOperation(operation)
			if mrErr != nil {
				log.Printf("[ERROR] Marshalling error: %s", mrErr)
			} else {
				mc <- *mr
			}
		}
	}(operationsChan, marshalledChan, waitGroup, format, fileHandler, errors)

	waitGroup.Add(1)
	go func(mc chan MarshalledResult, mu *sync.Mutex, wg *sync.WaitGroup, jsonFile *os.File, csvWriter *csv.Writer, format string, fileHandler IFileMarshaller, errors chan error) {
		defer func(iGroup *sync.WaitGroup, res chan bool) {
			iGroup.Done()
		}(wg, resultChan)

		for mr := range mc {
			fileHandler.WriteToFile(&mr)
		}
	}(marshalledChan, fileMutex, waitGroup, f, csvWriter, format, fileHandler, errors)

	waitGroup.Wait()
	// }(waitGroup)

	// //
	// <-resultChan
	// for md := range marshalledChan {
	// 	fmt.Println(md)
	// 	// operations = append(operations, operation)
	// 	// fmt.Println(operation)
	// }
	// fmt.Println(operations)
	// for rows.Next() {
	// 	operation := WalletOperation{}
	// 	scanErr := rows.Scan(&operation.ID, &operation.Operation, &operation.WalletFrom, &operation.WalletTo, &operation.Amount, &operation.CreatedAt)
	// 	if scanErr != nil {
	// 		fmt.Println("SCAN ERROR: ", scanErr)
	// 		return
	// 	}
	// 	if format == "json" {
	// jsonBytes, jsonMarshallErr := json.Marshal(operation)
	// if jsonMarshallErr != nil {
	// 	fmt.Println("SCAN ERROR: ", jsonMarshallErr)
	// 	return
	// }
	// newLine := []byte("\n")
	// _, writeErr := f.Write(jsonBytes)
	// f.Write(newLine)
	// if writeErr != nil {
	// 	fmt.Println("Write err: ", writeErr)
	// 	return
	// }
	// 	} else if format == "csv" && csvWriter != nil {
	// idStr := strconv.Itoa(operation.ID)
	// walletFromStr := strconv.Itoa(int(operation.WalletFrom.Int32))
	// walletToStr := strconv.Itoa(int(operation.WalletTo))
	// amountStr := operation.Amount.String()
	// createdAtStr := operation.CreatedAt.String()
	// row := []string{
	// 	idStr,
	// 	walletFromStr,
	// 	walletToStr,
	// 	amountStr,
	// 	createdAtStr,
	// }
	// 		csvWriter.Write(row)
	// 	}
	// }

	header := make([]byte, 512)
	f.Read(header)
	stat, _ := f.Stat()
	size := strconv.FormatInt(stat.Size(), 10)
	contentType := http.DetectContentType(header)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	f.Seek(0, 0)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", size)
	if csvWriter != nil {
		csvWriter.Flush()
	}
	http.ServeFile(w, r, fullPath)
}
