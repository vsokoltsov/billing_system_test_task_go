package operations

import (
	"billing_system_test_task/pkg/pipeline"
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

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
		waitGroup      = &sync.WaitGroup{}
		fileMutex      = &sync.Mutex{}
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
	readPipe := ReadPipe{
		or:     oh.OperationsRepo,
		oc:     operationsChan,
		wg:     waitGroup,
		params: listParams,
		ctx:    ctx,
	}
	MarshallPipe := MarshallPipe{
		wg: waitGroup,
		fm: fileHandler,
	}
	WritePipe := WritePipe{
		wg: waitGroup,
		fm: fileHandler,
	}
	pipes := []pipeline.Pipe{
		readPipe,
		MarshallPipe,
		WritePipe,
	}

	waitGroup.Add(3)
	pipeline.ExecutePipeline(pipes...)

	waitGroup.Wait()

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
