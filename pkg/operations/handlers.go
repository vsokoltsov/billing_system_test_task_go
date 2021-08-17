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
		format      = "json"
		csvWriter   *csv.Writer
		headers     []string
		f           *os.File
		fileOpenErr error
	)

	format = r.FormValue("format")
	if format == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
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

	rows, rowsErr := oh.OperationsRepo.List(ctx)
	if rowsErr != nil {
		fmt.Println(rowsErr)
		return
	}
	if format == "csv" {
		csvWriter = csv.NewWriter(f)
		// defer func(writer *csv.Writer) {
		// 	writer.Flush()
		// }(csvWriter)

		headers = []string{
			"id", "operation", "wallet_from", "wallet_to", "amount", "created_at",
		}
		csvWriter.Write(headers)
	}

	for rows.Next() {
		operation := WalletOperation{}
		scanErr := rows.Scan(&operation.ID, &operation.Operation, &operation.WalletFrom, &operation.WalletTo, &operation.Amount, &operation.CreatedAt)
		if scanErr != nil {
			fmt.Println("SCAN ERROR: ", scanErr)
			return
		}
		if format == "json" {
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
		} else if format == "csv" && csvWriter != nil {
			idStr := strconv.Itoa(operation.ID)
			walletFromStr := strconv.Itoa(int(operation.WalletFrom.Int32))
			walletToStr := strconv.Itoa(int(operation.WalletTo))
			amountStr := operation.Amount.String()
			createdAtStr := operation.CreatedAt.String()
			row := []string{
				idStr,
				walletFromStr,
				walletToStr,
				amountStr,
				createdAtStr,
			}
			csvWriter.Write(row)
		}
	}

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
