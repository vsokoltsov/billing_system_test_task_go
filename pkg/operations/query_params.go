package operations

import (
	"fmt"
	"net/url"
	"strconv"
)

// IQueryParamsReader represents actions for query parameters reading
type IQueryParamsReader interface {
	Parse(query url.Values) (*QueryParams, error)
}

// QueryParams represents parameters for
type QueryParams struct {
	format     string
	listParams *ListParams
}

// QueryParams implements IQueryParamsReader interface
type QueryParamsReader struct{}

// Parse returns given URL query parameters
func (qpr QueryParamsReader) Parse(query url.Values) (*QueryParams, error) {
	var (
		format string
		params *ListParams
	)
	format = query.Get("format")
	pageStr := query.Get("page")
	perPageStr := query.Get("per_page")

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

		params = &ListParams{
			page:    page,
			perPage: perPage,
		}
	}

	return &QueryParams{
		format:     format,
		listParams: params,
	}, nil
}
