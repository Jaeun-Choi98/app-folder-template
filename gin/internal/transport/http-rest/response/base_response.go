package response

type BaseResponse struct {
	Result int `json:"result"`
	Data   any `data:"data"`
}

var (
	// user define stats
	SUCCESS       = NewBaseResponse(0, nil)
	FAIL          = NewBaseResponse(1, nil)
	NACK          = NewBaseResponse(2, nil)
	TIMEOUT_SIG   = NewBaseResponse(3, nil)
	TIMEOUT_WEB   = NewBaseResponse(4, nil)
	INVAILD_DATA  = NewBaseResponse(5, nil)
	INVALID_TOKEN = NewBaseResponse(6, nil)

	INVAILD_DB_TYPE = NewBaseResponse(21, nil)

	INVALID_LOGIN_ID  = NewBaseResponse(100, nil)
	INVAILD_LOGIN_PWD = NewBaseResponse(101, nil)
)

func NewBaseResponse(result int, data any) *BaseResponse {
	return &BaseResponse{
		Result: result,
		Data:   data,
	}
}

func (r *BaseResponse) Add(data any) *BaseResponse {
	r.Data = data
	return r
}
