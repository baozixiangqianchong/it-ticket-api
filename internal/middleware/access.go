package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"it-ticket-api/internal/logger"
)

// 这些字段打日志时换成 ***，避免密码和 token 出现在终端。
var secretKeys = map[string]struct{}{
	"password":      {},
	"password_hash": {},
	"token":         {},
	"jwt":           {},
}

// AccessLog 请求结束后打一行：路径、参数、状态码、是否成功、耗时。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		params := readParams(c)

		c.Next() // 先执行后面的 handler，回来时才能拿到状态码

		status := c.Writer.Status()
		ok := status >= http.StatusOK && status < http.StatusBadRequest // 2xx / 3xx 算成功
		logger.Info("接口请求",
			"request_id", ID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"params", params,
			"status", status,
			"success", ok,
			"cost_ms", time.Since(start).Milliseconds(),
		)
	}
}

// readParams 拼出可打印的参数：query + JSON body。读完 body 要写回去，否则 handler 解不了 JSON。
func readParams(c *gin.Context) string {
	query := c.Request.URL.RawQuery

	var body string
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		raw, err := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw)) // 还原，后面 ShouldBindJSON 还能读
		if err == nil && len(raw) > 0 {
			body = maskJSON(raw)
		}
	}

	switch {
	case query != "" && body != "":
		return "query=" + query + " body=" + body
	case query != "":
		return "query=" + query
	case body != "":
		return body
	default:
		return "-"
	}
}

// maskJSON 把 JSON 里的密码、token 打码；解不开就原样截断，避免把乱码打很长。
func maskJSON(raw []byte) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		s := strings.TrimSpace(string(raw))
		if len(s) > 200 {
			return s[:200] + "..."
		}
		return s
	}
	for k := range m {
		if _, secret := secretKeys[strings.ToLower(k)]; secret {
			m[k] = "***"
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(out)
}
