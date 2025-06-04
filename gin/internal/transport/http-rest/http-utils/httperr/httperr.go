package httperr

import "net/http"

type HttpError struct {
	Code   int    `json:"code"`
	Msg    string `json:"message"`
	ErrMsg string `json:"error"`
}

const (
	UNAUTHORIZED_CODE = http.StatusNotAcceptable
	BADREQUEST_CODE   = http.StatusBadRequest
	INNER_ERROR_CODE  = http.StatusInternalServerError
)

var (
	UNAUTHORIZED = NewHttpError(UNAUTHORIZED_CODE, "권한 없음")
	BADREQUEST   = NewHttpError(BADREQUEST_CODE, "잘못된 요청")
	INNER_ERROR  = NewHttpError(INNER_ERROR_CODE, "서버 에러")
)

func NewHttpError(code int, msg string) *HttpError {
	return &HttpError{
		Code: code,
		Msg:  msg,
	}
}

func (h *HttpError) Error() string {
	return h.ErrMsg
}

func (h *HttpError) AddErrMsg(err error) *HttpError {
	if err != nil {
		h.ErrMsg = err.Error()
	}
	return h
}
