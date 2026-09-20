package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"it-ticket-api/internal/response"
)

// uidKey 是 context 里存放当前用户 id 的键。
// 中间件 c.Set，handler 用 UID() 取出。字符串写死一处，避免两边拼错。
const uidKey = "uid"

// JWT 返回一个 Gin 中间件：校验 Authorization: Bearer <token>。
// secret 来自配置 JWT_SECRET，必须和登录 signToken 用的同一把，否则合法 token 也会被拒。
// 失败直接 401 并 Abort，后面的 /me、提单等 handler 不会执行。
//
// 这里只回答「证件真不真」，不查库、不对比密码。查用户资料仍由 service 做。
func JWT(secret string) gin.HandlerFunc {
	// 闭包外面转成 []byte：每个请求都要拿去验 HMAC，避免在热路径里重复转换。
	key := []byte(secret)
	return func(c *gin.Context) {
		// GetHeader 读的是客户端请求头，不是 context。没带这个头就是没登录。
		raw := c.GetHeader("Authorization")
		if raw == "" {
			unauthorized(c)
			return
		}
		// 文档约定：Authorization: Bearer <token>。只写 token 或用自定义头（如 token: xxx）都会失败。
		const prefix = "Bearer "
		if !strings.HasPrefix(raw, prefix) {
			unauthorized(c)
			return
		}
		// 去掉前缀后应剩下 JWT 三段（头.载荷.签名）。只剩空格也当未登录。
		token := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		if token == "" {
			unauthorized(c)
			return
		}

		// Parse：拆开三段、用下面函数拿到密钥、验 HMAC；同时按 claims.exp 判断是否过期。
		t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			// 只接受 HS256。算法被改成 none 或不对称算法时拒绝，避免换算法绕过验签。
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		})
		// 签名错、过期、格式坏了：统一 401，不告诉客户端具体是哪一种。
		if err != nil || t == nil || !t.Valid {
			unauthorized(c)
			return
		}
		// Claims 是接口；断言成 MapClaims 后才能按键取 uid / email / role。
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			unauthorized(c)
			return
		}
		uid, ok := uidFromClaims(claims)
		if !ok {
			unauthorized(c)
			return
		}
		// 写入本次请求的 context。handler 用 UID(c) 取，不要再自己解析 JWT。
		c.Set(uidKey, uid)
		c.Next() // 放行，进入后面的 /me、POST /tickets 等
	}
}

// UID 取出中间件写入的当前用户 id。没经过 JWT 或值不是正整数则 ok=false。
func UID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(uidKey)
	if !ok {
		return 0, false
	}
	// context 存的是 any，必须断言回 int64，和 c.Set 放进去的类型一致。
	id, ok := v.(int64)
	return id, ok && id > 0
}

func unauthorized(c *gin.Context) {
	response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
	c.Abort() // 必须 Abort：只 Fail 不中止的话，后面的 handler 还会继续跑
}

// uidFromClaims 取出签发时写入的 uid。
// jwt.Parse 按 JSON 解数字，得到的是 float64，不能直接写成 claims["uid"].(int64)。
func uidFromClaims(claims jwt.MapClaims) (int64, bool) {
	raw, ok := claims["uid"]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		id := int64(v)
		// 1.5 转成 int64 会变成 1，再转回 float 对不上，用来拒绝「不是整数的 uid」。
		if float64(id) != v || id <= 0 {
			return 0, false
		}
		return id, true
	case int64:
		return v, v > 0
	default:
		return 0, false
	}
}
