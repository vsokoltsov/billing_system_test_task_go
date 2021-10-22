package usecases

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/repositories"
	"billing_system_test_task/internal/repositories/reports"
	"context"
	"net/url"
)

type WalletOperationUsecase interface {
	GenerateReport(ctx context.Context, queryParams url.Values) (*entities.FileMetadata, adapters.Error)
}

type WalletOperationInteractor struct {
	walletOperationRepo     repositories.OperationsManager
	queryParameters         reports.QueryReaderManager
	fileHandler             reports.FileHandlingManager
	operationProcessManager reports.PipelineManager
	errorsFactory           adapters.ErrorsFactory
}

func NewWalletOperationInteractor(walletOperationRepo repositories.OperationsManager, queryParameters reports.QueryReaderManager, fileHandler reports.FileHandlingManager, operationProcessManager reports.PipelineManager, errorsFactory adapters.ErrorsFactory) *WalletOperationInteractor {
	return &WalletOperationInteractor{
		walletOperationRepo:     walletOperationRepo,
		queryParameters:         queryParameters,
		fileHandler:             fileHandler,
		operationProcessManager: operationProcessManager,
		errorsFactory:           errorsFactory,
	}
}

func (wor *WalletOperationInteractor) GenerateReport(ctx context.Context, queryParams url.Values) (*entities.FileMetadata, adapters.Error) {
	// Parse query parameters
	qp, qpErr := wor.queryParameters.Parse(queryParams)
	if qpErr != nil {
		return nil, wor.errorsFactory.DefaultError(qpErr)
	}

	// Create new file for report
	fileParams, fpErr := wor.fileHandler.Create(qp.Format)
	if fpErr != nil {
		return nil, wor.errorsFactory.DefaultError(fpErr)
	}

	// Creates file marshaller
	fileHandler, fhErr := wor.fileHandler.CreateMarshaller(
		fileParams.File,
		qp.Format,
		fileParams.CsvWriter,
	)
	if fhErr != nil {
		return nil, wor.errorsFactory.DefaultError(fhErr)
	}

	// Process receiving, marshalling and writing to file wallet operations
	processErr := wor.operationProcessManager.Process(ctx, wor.walletOperationRepo, qp.ListParams, fileHandler)
	if processErr != nil {
		return nil, wor.errorsFactory.DefaultError(processErr)
	}

	// Get file metadata
	metadata, metadataErr := wor.fileHandler.GetFileMetadata(fileParams.File)
	if metadataErr != nil {
		return nil, wor.errorsFactory.DefaultError(metadataErr)
	}
	return &entities.FileMetadata{
		File:        fileParams.File,
		Path:        fileParams.Path,
		Size:        metadata.Size,
		ContentType: metadata.ContentType,
	}, nil
}
