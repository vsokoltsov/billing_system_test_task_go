package reports

import (
	"billing_system_test_task/pkg/repositories"
	"fmt"
	"net/url"
	"strconv"
)

// QueryReaderManager represents actions for query parameters reading
type QueryReaderManager interface {
	Parse(query url.Values) (*QueryParams, error)
}

// QueryParams represents parameters for
type QueryParams struct {
	Format     string
	ListParams *repositories.ListParams
}

// QueryParams implements QueryReaderManager interface
type QueryParamsReader struct{}

func NewQueryParamsReader() *QueryParamsReader {
	return &QueryParamsReader{}
}

// Parse returns given URL query parameters
func (qpr QueryParamsReader) Parse(query url.Values) (*QueryParams, error) {
	var (
		format string
		params = &repositories.ListParams{}
	)
	format = query.Get("format")
	pageStr := query.Get("page")
	perPageStr := query.Get("per_page")
	date := query.Get("date")

	if format == "" {
		format = "json"
	}

	if pageStr != "" && perPageStr != "" {
		page, pageConvError := strconv.Atoi(pageStr)
		if pageConvError != nil {
			return nil, fmt.Errorf("error of 'page' attribute converting: %s", pageConvError)
		}

		perPage, perPageConvError := strconv.Atoi(perPageStr)
		if perPageConvError != nil {
			return nil, fmt.Errorf("error of 'per_page' attribute converting: %s", pageConvError)
		}

		params.Page = page
		params.PerPage = perPage
	}
	if date != "" {
		params.Date = date
	}

	return &QueryParams{
		Format:     format,
		ListParams: params,
	}, nil
}
