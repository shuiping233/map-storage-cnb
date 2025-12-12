package model

// 最里层：仅放业务数据
type DataAny interface{}

// 通用外壳
type CommonResp struct {
	Msg  string  `json:"msg"`
	Data DataAny `json:"data,omitempty"` // omitempty 可以让 data==nil 时字段消失
}

// 快速构造成功响应
func OK(data DataAny) CommonResp {
	return CommonResp{Msg: "success", Data: data}
}

// 快速构造成功响应
func Common(msg string) CommonResp {
	return CommonResp{Msg: msg}
}

// 快速构造错误响应
func Fail(msg string) CommonResp {
	return CommonResp{Msg: msg}
}

// 快速构造错误响应
func FailWithData(msg string, data DataAny) CommonResp {
	return CommonResp{Msg: msg, Data: data}
}
