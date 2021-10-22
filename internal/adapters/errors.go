package adapters

type Error interface {
	GetError() error
	GetStatus() int
}

type ErrorsFactory interface {
	NotFound(err error) Error
	DefaultError(err error) Error
}

type HTTPErrorsFactory struct{}

func NewHTTPErrorsFactory() *HTTPErrorsFactory {
	return &HTTPErrorsFactory{}
}

func (he *HTTPErrorsFactory) NotFound(err error) Error {
	return NewHTTPError(
		404, err,
	)
}

func (he *HTTPErrorsFactory) DefaultError(err error) Error {
	return NewHTTPError(
		400, err,
	)
}

type HTTPError struct {
	status int
	err    error
}

func NewHTTPError(status int, err error) *HTTPError {
	return &HTTPError{
		status: status,
		err:    err,
	}
}

func (err *HTTPError) GetError() error {
	return err.err
}

func (err *HTTPError) GetStatus() int {
	return err.status
}
