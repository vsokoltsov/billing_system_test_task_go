package http

import (
	"billing_system_test_task/pkg/usecases"
	"billing_system_test_task/pkg/utils"
	"context"
	"net/http"
	"os"
)

// OperationsHandler represents handler structure for the operatons
type OperationsHandler struct {
	woUseCase usecases.WalletOperationUsecase
}

// NewOperationsHandler returns controller instance
func NewOperationsHandler(woUseCase usecases.WalletOperationUsecase) OperationsHandler {
	return OperationsHandler{
		woUseCase: woUseCase,
	}
}

// Create godoc
// @Summary Wallet operations
// @Description Get wallet operations logs
// @Tags operations
// @Accept  json
// @Produce application/octet-stream
// @Param format query string false "Report format"
// @Param page query int false "Page number"
// @Param per_page query int false "Number of items per page"
// @Param date query int false "Number of items per page"
// @Router /api/operations/ [get]
// @Header 200 {string} Content-Type "application/octet-stream"
// @Header 200 {string} Expires "0"
func (oh *OperationsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	v := r.URL.Query()
	fileMetadata, grErr := oh.woUseCase.GenerateReport(ctx, v)
	if grErr != nil {
		utils.JsonResponseError(w, grErr.GetStatus(), grErr.GetError().Error())
		return
	}
	// queryParams, qpErr := oh.pr.Parse(v)
	// if qpErr != nil {
	// utils.JsonResponseError(w, http.StatusBadRequest, qpErr.Error())
	// return
	// }

	// fileParams, fileCreateErr := oh.fh.Create(queryParams.format)
	// if fileCreateErr != nil {
	// 	utils.JsonResponseError(w, http.StatusBadRequest, fileCreateErr.Error())
	// 	return
	// }

	defer func(path string, f *os.File) {
		f.Close()
		os.Remove(path)
	}(fileMetadata.Path, fileMetadata.File)

	// fileHandler, fileHandlerErr := oh.fh.CreateMarshaller(
	// fileParams.f,
	// queryParams.format,
	// fileParams.csvWriter,
	// )
	// if fileHandlerErr != nil {
	// 	utils.JsonResponseError(w, http.StatusBadRequest, fileHandlerErr.Error())
	// 	return
	// }

	// processErr := oh.op.Process(ctx, oh.or, queryParams.listParams, fileHandler)
	// if processErr != nil {
	// 	utils.JsonResponseError(w, http.StatusBadRequest, processErr.Error())
	// 	return
	// }

	// metadata, metadataErr := oh.fh.GetFileMetadata(fileParams.f)
	// if metadataErr != nil {
	// 	utils.JsonResponseError(w, http.StatusBadRequest, metadataErr.Error())
	// 	return
	// }
	w.Header().Set("Content-Disposition", "attachment; filename="+fileMetadata.File.Name())
	w.Header().Set("Content-Type", fileMetadata.ContentType)
	w.Header().Set("Content-Length", fileMetadata.Size)
	w.WriteHeader(http.StatusOK)
	// if fileParams.csvWriter != nil {
	// 	fileParams.csvWriter.Flush()
	// }

	http.ServeFile(w, r, fileMetadata.Path)
}
