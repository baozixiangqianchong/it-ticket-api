package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务码：成功用 0；失败用和 HTTP 对应的数字，方便前端判断。
const (
	CodeOK              = 0   // 成功
	CodeInvalidArg      = 400 // 参数错误（JSON 不合法、缺字段、密码太短等）
	CodeUnauthenticated = 401 // 未登录或登录失败（token 无效、邮箱或密码错误）
	CodeForbidden       = 403 // 已登录但权限不够（例如员工去指派工单）
	CodeNotFound        = 404 // 资源不存在，或存在但对当前用户不可见
	CodeConflict        = 409 // 冲突（邮箱已被注册、工单状态不允许这样跳）
	CodeInternal        = 500 // 服务器内部错误
	CodeUnavailable     = 503 // 依赖不可用（健康检查发现 MySQL 不通）
)

// Body 所有接口共用的外层信封。code 是数字，message 给人看（中文）。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func OK(c *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(http.StatusOK, Body{Code: CodeOK, Message: "成功", Data: data})
}

func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}
