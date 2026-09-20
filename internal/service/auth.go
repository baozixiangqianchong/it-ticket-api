package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

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
			return nil, emailTaken()
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

// Me 用登录签发的 JWT 换当前用户公开资料。token 无效、过期或用户已不存在都统一 401。
func (s *AuthService) Me(token string) (*model.PublicUser, error) {
	// 中间件已经验过，这里再剥一次 Bearer，方便 handler 把头原样传进来。
	token = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(token), "Bearer "))
	if token == "" {
		return nil, tokenInvalid()
	}
	claims, err := s.verifyToken(token)
	if err != nil {
		return nil, err
	}
	uid, ok := uidFromClaims(claims)
	if !ok {
		return nil, tokenInvalid()
	}
	// 以库为准：token 里没有 display_name，改名、改角色后也能看到新值。
	u, err := s.users.FindByID(uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, tokenInvalid() // token 还有效，但人已经不在了
		}
		return nil, err
	}
	pub := u.Public()
	return &pub, nil
}

// verifyToken 验签并检查过期。密钥必须和 signToken 用的同一把 JWT_SECRET。
// 成功返回 claims（里面有 uid / email / role）；失败统一 401，不区分「过期」还是「伪造」。
func (s *AuthService) verifyToken(token string) (jwt.MapClaims, error) {
	// Parse 做三件事：拆开 JWT 三段、用下面这个函数拿到密钥、用密钥验 HMAC。
	// 第二个参数是「钥匙从哪来」：库验签时会调用它，我们把登录时用的 jwtSecret 交出去。
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		// 只接受 HS256。有人改成 none 或不对称算法，这里直接拒绝，避免被换算法绕过验签。
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	// 签名不对、过期（claims 里的 exp）、格式坏了，都会走进这里。
	if err != nil || t == nil || !t.Valid {
		return nil, tokenInvalid()
	}
	// t.Claims 的静态类型是接口 jwt.Claims，里面实际是 Parse 解出来的 payload。
	// 断言成 MapClaims（本质是 map[string]any）之后，才能用 claims["uid"] 取值。
	// 和 signToken 写入的是同一份字段，例如：
	//   uid=2, email=alice@example.com, role=user, exp=过期时间, iat=签发时间
	// ok=false：payload 不是 map（极少见），当 token 无效处理。
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, tokenInvalid()
	}
	return claims, nil
}

// uidFromClaims 取出签发时写入的 uid。JSON 数字解析后是 float64，不能直接断言成 int64。
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
