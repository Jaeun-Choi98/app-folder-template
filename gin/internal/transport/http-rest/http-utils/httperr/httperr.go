package httperr

import (
	"net/http"
	"pjt/internal/transport/http-rest/response"
)

type HttpError struct {
	Code        int
	ErrResponse *response.BaseResponse
	ErrMsg      string
}

const (
	UNAUTHORIZED_CODE = http.StatusNotAcceptable
	BADREQUEST_CODE   = http.StatusBadRequest
	INNER_ERROR_CODE  = http.StatusInternalServerError
)

var (
	// http status
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
	if err != nil {
		h.ErrMsg = err.Error()
	}
	if res != nil {
		h.ErrResponse = res
	}
	return h
}
