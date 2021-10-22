package http

import (
	"billing_system_test_task/internal/usecases"
	"context"
	"net/http"
	"os"
)

// OperationsHandler represents handler structure for the operatons
type OperationsHandler struct {
	woUseCase usecases.WalletOperationUsecase
}

// NewOperationsHandler returns controller instance
func NewOperationsHandler(woUseCase usecases.WalletOperationUsecase) *OperationsHandler {
	return &OperationsHandler{
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
		JsonResponseError(w, grErr.GetStatus(), grErr.GetError().Error())
		return
	}

	defer func(path string, f *os.File) {
		f.Close()
		os.Remove(path)
	}(fileMetadata.Path, fileMetadata.File)

	w.Header().Set("Content-Disposition", "attachment; filename="+fileMetadata.File.Name())
	w.Header().Set("Content-Type", fileMetadata.ContentType)
	w.Header().Set("Content-Length", fileMetadata.Size)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, fileMetadata.Path)
}
