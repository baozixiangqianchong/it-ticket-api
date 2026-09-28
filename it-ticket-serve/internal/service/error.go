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
	Name    string // 响应里的 error 字符串，用来区分同状态码的冲突
	Message string // 给前端看的中文说明
}

func (e *Error) Error() string { return e.Message }

func apiErr(status, code int, name, msg string) *Error {
	return &Error{Status: status, Code: code, Name: name, Message: msg}
}

func invalidArg(msg string) *Error {
	return apiErr(http.StatusBadRequest, response.CodeInvalidArg, response.ErrInvalidArgument, msg)
}

func unauthenticated() *Error {
	return apiErr(http.StatusUnauthorized, response.CodeUnauthenticated, response.ErrUnauthenticated, "邮箱或密码错误")
}

func tokenInvalid() *Error {
	return apiErr(http.StatusUnauthorized, response.CodeUnauthenticated, response.ErrUnauthenticated, "未认证")
}

func emailTaken() *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrEmailTaken, "该邮箱已被注册")
}

func notFound(msg string) *Error {
	return apiErr(http.StatusNotFound, response.CodeNotFound, response.ErrNotFound, msg)
}

func permissionDenied(msg string) *Error {
	return apiErr(http.StatusForbidden, response.CodeForbidden, response.ErrPermissionDenied, msg)
}

func ticketInvalidTransition(msg string) *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrTicketInvalidTrans, msg)
}

func ticketClosed(msg string) *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrTicketClosed, msg)
}

func ticketAlreadyAssigned(msg string) *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrTicketAlreadyAssigned, msg)
}

func lastAdmin() *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrLastAdmin, "不能取消最后一个管理员")
}

func lastActiveAdmin() *Error {
	return apiErr(http.StatusConflict, response.CodeConflict, response.ErrLastAdmin, "不能停用最后一个管理员")
}

func accountDisabled() *Error {
	return apiErr(http.StatusUnauthorized, response.CodeUnauthenticated, response.ErrUnauthenticated, "账号已停用")
}
