package response

type Response interface {
	SetError(code int, err error, content string)
	SetResponse(data interface{})
}

/*
消息返回响应结构体
Code: 响应码, == 0 成功, < 0 失败
Error: 失败名称
Content: 成功/失败描述
ResponseID:
Data: 响应数据
*/
type ResponseResult struct {
	Response   `json:"-,omitempty"`
	Code       int         `json:"code"`
	Error      string      `json:"error"`
	Content    string      `json:"contnet"`
	ResponseID string      `json:"responseid"`
	Data       interface{} `json:"data,omitpty"`
}

func (r *ResponseResult) SetError(code int, err error, content string) {

	r.Code = code
	r.Error = err.Error()
	r.Content = content
}

func (r *ResponseResult) SetResponse(data interface{}) {

	r.Data = data
}
