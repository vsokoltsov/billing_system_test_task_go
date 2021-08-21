package operations

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

type IFilehandler interface {
	MarshallOperation(operation *WalletOperation) (*MarshalledResult, error)
	WriteToFile(mr *MarshalledResult) error
}

type JSONHandler struct {
	file *os.File
	mu   *sync.Mutex
}

func (jh *JSONHandler) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
	jsonBytes, jsonMarshallErr := json.Marshal(operation)
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
	bytesData := mr.data.([]byte)
	jh.mu.Lock()
	jh.file.Sync()
	jh.file.Write(bytesData)
	jh.file.Sync()
	jh.mu.Unlock()
	return nil
}

type CSVHandler struct {
	csvWriter *csv.Writer
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
	ch.csvWriter.Write(row)
	ch.mu.Unlock()
	return nil
}

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
		// readJobsChn    = make(<-chan ReadJob)
		// operations     []*WalletOperation
		fileMutex   = &sync.Mutex{}
		resultChan  = make(chan bool)
		fileHandler IFilehandler
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
	go func(handler *OperationsHandler, oc chan *WalletOperation, wg *sync.WaitGroup) {
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
				fmt.Println("SCAN ERROR: ", scanErr)
				return
			}
			oc <- &operation
		}

	}(oh, operationsChan, waitGroup)

	waitGroup.Add(1)
	go func(oc chan *WalletOperation, mc chan MarshalledResult, wg *sync.WaitGroup, format string, fileHandler IFilehandler) {
		defer close(mc)
		defer wg.Done()

		for operation := range operationsChan {
			mr, _ := fileHandler.MarshallOperation(operation)
			mc <- *mr
		}
	}(operationsChan, marshalledChan, waitGroup, format, fileHandler)

	waitGroup.Add(1)
	go func(mc chan MarshalledResult, result chan bool, mu *sync.Mutex, wg *sync.WaitGroup, jsonFile *os.File, csvWriter *csv.Writer, format string, fileHandler IFilehandler) {
		defer func(iGroup *sync.WaitGroup, res chan bool) {
			iGroup.Done()
			res <- true
		}(wg, resultChan)

		for mr := range mc {
			fileHandler.WriteToFile(&mr)
		}
	}(marshalledChan, resultChan, fileMutex, waitGroup, f, csvWriter, format, fileHandler)

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
