package response

type BaseResponse struct {
	Result int `json:"result"`
	Data   any `data:"data"`
}

var (
	// user define stats
	SUCCESS       = NewBaseResponse(0)
	FAIL          = NewBaseResponse(1)
	NACK          = NewBaseResponse(2)
	TIMEOUT_SIG   = NewBaseResponse(3)
	TIMEOUT_WEB   = NewBaseResponse(4)
	INVAILD_DATA  = NewBaseResponse(5)
	INVALID_TOKEN = NewBaseResponse(6)

	INVAILD_DB_TYPE = NewBaseResponse(21)

	INVALID_LOGIN_ID  = NewBaseResponse(100)
	INVAILD_LOGIN_PWD = NewBaseResponse(101)
)

func NewBaseResponse(result int) *BaseResponse {
	return &BaseResponse{
		Result: result,
	}
}

func (r *BaseResponse) Add(data any) *BaseResponse {
	n := NewBaseResponse(r.Result)
	n.Data = data
	return n
}
