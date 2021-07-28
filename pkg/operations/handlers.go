package operations

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	f, err := os.OpenFile(filepath.Join(basepath, "report.txt"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	rows, rowsErr := oh.OperationsRepo.List(ctx)
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
		jsonBytes, jsonMarshallErr := json.Marshal(operation)
		if jsonMarshallErr != nil {
			fmt.Println("SCAN ERROR: ", jsonMarshallErr)
			return
		}
		newLine := []byte("\n")
		_, writeErr := f.Write(jsonBytes)
		f.Write(newLine)
		if writeErr != nil {
			fmt.Println("Write err: ", writeErr)
			return
		}
	}

	header := make([]byte, 512)
	f.Read(header)
	stat, _ := f.Stat()
	size := strconv.FormatInt(stat.Size(), 10)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	f.Seek(0, 0)
	w.Header().Set("Content-Disposition", "attachment; filename="+"report.txt")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", size)

	io.Copy(w, f)
}
