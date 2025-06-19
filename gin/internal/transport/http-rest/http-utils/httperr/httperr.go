package httperr

import "net/http"

type HttpError struct {
	Code        int
	ErrResponse *ErrResponse
	ErrMsg      string
}

type ErrResponse struct {
	Result int `json:"result"`
	Data   any `data:"data"`
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

	// user define stats
	FAIL          = NewErrResponse(1, nil)
	NACK          = NewErrResponse(2, nil)
	TIMEOUT_SIG   = NewErrResponse(3, nil)
	TIMEOUT_WEB   = NewErrResponse(4, nil)
	INVAILD_DATA  = NewErrResponse(5, nil)
	INVALID_TOKEN = NewErrResponse(6, nil)

	INVAILD_DB_TYPE = NewErrResponse(21, nil)

	INVALID_LOGIN_ID  = NewErrResponse(100, nil)
	INVAILD_LOGIN_PWD = NewErrResponse(101, nil)
)

func NewHttpError(code int) *HttpError {
	return &HttpError{
		Code: code,
	}
}

func NewErrResponse(result int, data any) *ErrResponse {
	return &ErrResponse{
		Result: result,
		Data:   data,
	}
}

func (h *HttpError) Error() string {
	return h.ErrMsg
}

func (h *HttpError) Add(err error, res *ErrResponse) *HttpError {
	if err != nil {
		h.ErrMsg = err.Error()
	}
	if res != nil {
		h.ErrResponse = res
	}
	return h
}

func (r *ErrResponse) Add(data any) *ErrResponse {
	r.Data = data
	return r
}
