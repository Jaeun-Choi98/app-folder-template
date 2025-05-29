package httperr

import "net/http"

type HttpError struct {
	Code   int    `json:"code"`
	Msg    string `json:"message"`
	ErrMsg string `json:"error"`
}

var (
	UNAUTHORIZED = NewHttpError(http.StatusNotAcceptable, "권한 없음")
	BADREQUEST   = NewHttpError(http.StatusBadRequest, "잘못된 요청")
	INNER_ERROR  = NewHttpError(http.StatusInternalServerError, "서버 에러")
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
