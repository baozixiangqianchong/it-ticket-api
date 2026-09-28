package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/response"
)

const uidKey = "uid"
const roleKey = "role"

// LoadRole 用 token 里的 uid 查库。found=false 表示人已经不在了。
type LoadRole func(uid int64) (role string, disabled bool, found bool, err error)

// JWT 只验「是谁」。角色每次查库，改角色后下一请求立刻生效，不采信 token 里的 role。
func JWT(secret string, loadRole LoadRole) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			unauthorized(c)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(raw, prefix) {
			unauthorized(c)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		if token == "" {
			unauthorized(c)
			return
		}

		t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		})
		if err != nil || t == nil || !t.Valid {
			unauthorized(c)
			return
		}
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

		role, disabled, found, err := loadRole(uid)
		if err != nil {
			logger.Error("加载用户角色失败", "uid", uid, "err", err)
			response.Fail(c, http.StatusInternalServerError, response.CodeInternal, "服务器内部错误")
			c.Abort()
			return
		}
		if !found || (role != "user" && role != "agent" && role != "admin") {
			unauthorized(c)
			return
		}
		if disabled {
			response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "账号已停用")
			c.Abort()
			return
		}
		if tokenRole, ok := roleFromClaims(claims); ok && tokenRole != role {
			logger.Info("角色已变更，以库为准", "uid", uid, "token_role", tokenRole, "db_role", role)
		}

		c.Set(uidKey, uid)
		c.Set(roleKey, role)
		c.Next()
	}
}

func UID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(uidKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok && id > 0
}

func Role(c *gin.Context) (string, bool) {
	v, ok := c.Get(roleKey)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	if !ok || (role != "user" && role != "agent" && role != "admin") {
		return "", false
	}
	return role, true
}

func unauthorized(c *gin.Context) {
	response.Fail(c, http.StatusUnauthorized, response.CodeUnauthenticated, "未认证")
	c.Abort()
}

func uidFromClaims(claims jwt.MapClaims) (int64, bool) {
	raw, ok := claims["uid"]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		id := int64(v)
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

func roleFromClaims(claims jwt.MapClaims) (string, bool) {
	raw, ok := claims["role"]
	if !ok {
		return "", false
	}
	role, ok := raw.(string)
	if !ok {
		return "", false
	}
	switch role {
	case "user", "agent", "admin":
		return role, true
	default:
		return "", false
	}
}
