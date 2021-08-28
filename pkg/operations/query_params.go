package operations

import (
	"fmt"
	"net/url"
	"strconv"
)

type IQueryParamsReader interface {
	Parse(query url.Values) (*QueryParams, error)
}

type QueryParams struct {
	format     string
	listParams *ListParams
}

type QueryParamsReader struct{}

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
