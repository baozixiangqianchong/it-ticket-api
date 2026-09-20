package service

import (
	"net/http"

	"it-ticket-api/internal/response"
)

// Error 业务错误。handler 读 Status/Code 写 HTTP 响应，不把 SQL 细节返回给客户端。
// 账号、工单共用这一份，不要在每个 service 文件里再定义一遍。
type Error struct {
	Status  int    // HTTP 状态码，如 400、401、409
	Code    int    // 响应里的数字业务码，和 HTTP 对齐
	Message string // 给前端看的中文说明
}

func (e *Error) Error() string { return e.Message }

func invalidArg(msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: response.CodeInvalidArg, Message: msg}
}

func unauthenticated() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: response.CodeUnauthenticated, Message: "邮箱或密码错误"}
}

func tokenInvalid() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: response.CodeUnauthenticated, Message: "未认证"}
}

func emailTaken() *Error {
	return &Error{Status: http.StatusConflict, Code: response.CodeConflict, Message: "该邮箱已被注册"}
}

func notFound(msg string) *Error {
	return &Error{Status: http.StatusNotFound, Code: response.CodeNotFound, Message: msg}
}
