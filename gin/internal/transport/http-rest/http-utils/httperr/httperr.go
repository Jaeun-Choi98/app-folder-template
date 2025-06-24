package httperr

import (
	"net/http"
	"pjt/internal/transport/http-rest/response"
)

type HttpError struct {
	Code         int
	BaseResponse *response.BaseResponse
	ErrMsg       string
}

const (
	NOT_FOUND_CODE    = http.StatusNotFound
	UNAUTHORIZED_CODE = http.StatusNotAcceptable
	BADREQUEST_CODE   = http.StatusBadRequest
	INNER_ERROR_CODE  = http.StatusInternalServerError
)

var (
	// http status
	NOT_FOUND    = NewHttpError(NOT_FOUND_CODE)
	UNAUTHORIZED = NewHttpError(UNAUTHORIZED_CODE)
	BADREQUEST   = NewHttpError(BADREQUEST_CODE)
	INNER_ERROR  = NewHttpError(INNER_ERROR_CODE)
)

func NewHttpError(code int) *HttpError {
	return &HttpError{
		Code: code,
	}
}

func (h *HttpError) Error() string {
	return h.ErrMsg
}

func (h *HttpError) Add(err error, res *response.BaseResponse) *HttpError {
	n := NewHttpError(h.Code)
	if err != nil {
		n.ErrMsg = err.Error()
	}
	n.BaseResponse = res
	return n
}
