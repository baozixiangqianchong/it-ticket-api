package service

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
	"it-ticket-api/internal/store"
)

// Error 业务错误。handler 读 Status/Code 写 HTTP 响应，不把 SQL 细节返回给客户端。
type Error struct {
	Status  int    // HTTP 状态码，如 400、401、409
	Code    int    // 响应里的数字业务码，和 HTTP 对齐
	Message string // 给前端看的中文说明
}

func (e *Error) Error() string { return e.Message }

// AuthService 注册、登录。密码哈希和签发 JWT 都在这里，不进 handler、不进 store。
type AuthService struct {
	users     *store.UserStore // 查/写 users 表
	jwtSecret []byte           // 签名密钥，来自配置 JWT_SECRET
}

func NewAuthService(users *store.UserStore, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret)}
}

// Register 创建普通员工。客户端就算传 role 也没用，这里写死 "user"。
func (s *AuthService) Register(in model.RegisterInput) (*model.PublicUser, error) {
	email, displayName, err := validateRegister(in)
	if err != nil {
		return nil, err
	}
	// bcrypt 单向哈希，库里只存 hash，不能反推出明文。DefaultCost 是计算强度。
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 角色固定 user：管理员只能 bootstrap 或以后由 admin 接口提升，不能靠注册自封。
	u, err := s.users.Create(email, string(hash), displayName, "user")
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			return nil, &Error{
				Status:  http.StatusConflict,
				Code:    response.CodeConflict,
				Message: "该邮箱已被注册",
			}
		}
		return nil, err
	}
	pub := u.Public() // 去掉 password_hash，避免返回给前端
	return &pub, nil
}

// Login 校验邮箱密码，成功则发 JWT。失败统一 401，不告诉对方「用户不存在」还是「密码错」。
func (s *AuthService) Login(in model.LoginInput) (*model.LoginResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" || in.Password == "" {
		return nil, unauthenticated()
	}
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	// u == nil：没有这个邮箱。CompareHash 失败：密码不对。两条路同一错误，防止被枚举账号。
	if u == nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		return nil, unauthenticated()
	}
	token, err := s.signToken(u)
	if err != nil {
		return nil, err
	}
	return &model.LoginResult{Token: token, User: u.Public()}, nil
}

// signToken 签发 JWT。后续请求头带 Authorization: Bearer <token>，中间件用同一密钥验签。
func (s *AuthService) signToken(u *model.User) (string, error) {
	// MapClaims 就是 JWT 里那一段 JSON。uid/role 给后面鉴权用，不必再查库才能知道是谁。
	claims := jwt.MapClaims{
		"uid":   u.ID,
		"email": u.Email,
		"role":  u.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(), // 24 小时后过期
		"iat":   time.Now().Unix(),                     // 签发时间
	}
	// HS256：用密钥做 HMAC。密钥泄露则任何人都能伪造 token。
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}

// validateRegister 校验并整理入参。email 转小写，避免 Alice@x.com 和 alice@x.com 当成两个账号。
func validateRegister(in model.RegisterInput) (email, displayName string, err error) {
	email = strings.TrimSpace(strings.ToLower(in.Email))
	displayName = strings.TrimSpace(in.DisplayName)
	if !validEmail(email) {
		return "", "", invalidArg("邮箱格式不正确")
	}
	// bcrypt 最多处理 72 字节；文档要求最短 8。用 len 按字节，和 bcrypt 限制一致。
	n := len(in.Password)
	if n < 8 || n > 72 {
		return "", "", invalidArg("密码长度需为 8～72 位")
	}
	// 显示名按「字」计数，中文「张三」是 2，不会被当成 6 个字节超限。
	if displayName == "" || utf8.RuneCountInString(displayName) > 64 {
		return "", "", invalidArg("显示名必填，最多 64 个字")
	}
	return email, displayName, nil
}

// validEmail 只做第一期要求的「基本格式」：有 @，@ 前后都有内容，域名里有点。
func validEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	return strings.Contains(s[at+1:], ".")
}

func invalidArg(msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: response.CodeInvalidArg, Message: msg}
}

func unauthenticated() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: response.CodeUnauthenticated, Message: "邮箱或密码错误"}
}
