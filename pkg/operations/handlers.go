package operations

import (
	"billing_system_test_task/pkg/utils"
	"context"
	"net/http"
	"os"
	"strconv"
)

type OperationsHandler struct {
	or IWalletOperationRepo
	op IOperationsProcessor
	pr IQueryParamsReader
	fh IFileHandling
}

func NewOperationsHandler(or IWalletOperationRepo, pr IQueryParamsReader, fh IFileHandling, op IOperationsProcessor) OperationsHandler {
	return OperationsHandler{
		or: or,
		pr: pr,
		fh: fh,
		op: op,
	}
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

	v := r.URL.Query()
	queryParams, qpErr := oh.pr.Parse(v)
	if qpErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, qpErr.Error())
		return
	}

	fileParams, fileCreateErr := oh.fh.Create(queryParams.format)
	if fileCreateErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, fileCreateErr.Error())
		return
	}

	defer func(path string, f *os.File) {
		f.Close()
		os.Remove(path)
	}(fileParams.path, fileParams.f)

	fileHandler := oh.fh.CreateMarshaller(
		fileParams.f,
		queryParams.format,
		fileParams.csvWriter,
	)

	processErr := oh.op.Process(ctx, oh.or, queryParams.listParams, fileHandler)
	if processErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, processErr.Error())
		return
	}

	header := make([]byte, 512)
	fileParams.f.Read(header)
	stat, _ := fileParams.f.Stat()
	size := strconv.FormatInt(stat.Size(), 10)
	contentType := http.DetectContentType(header)
	fileParams.f.Seek(0, 0)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileParams.name)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", size)
	w.WriteHeader(http.StatusOK)
	if fileParams.csvWriter != nil {
		fileParams.csvWriter.Flush()
	}

	http.ServeFile(w, r, fileParams.path)
}
