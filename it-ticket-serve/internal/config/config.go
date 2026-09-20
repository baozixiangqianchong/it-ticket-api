package config

import (
	"fmt"
	"os"
	"strings"
)

// Config 启动参数，全部来自环境变量，代码里不写死数据库密码。
type Config struct {
	HTTPAddr       string
	MySQLDSN       string
	JWTSecret      string
	BootstrapEmail string
	BootstrapPass  string
}

func Load() Config {
	httpAddr := getenv("HTTP_ADDR", ":8080")
	jwtSecret := getenv("JWT_SECRET", "dev-jwt-secret")
	bootEmail := getenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")
	bootPass := getenv("BOOTSTRAP_ADMIN_PASSWORD", "password1")

	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		return Config{
			HTTPAddr:       httpAddr,
			MySQLDSN:       dsn,
			JWTSecret:      jwtSecret,
			BootstrapEmail: bootEmail,
			BootstrapPass:  bootPass,
		}
	}

	user := getenv("MYSQL_USER", "root")
	pass := getenv("MYSQL_PASSWORD", "")
	host := getenv("MYSQL_HOST", "127.0.0.1")
	port := getenv("MYSQL_PORT", "3306")
	db := getenv("MYSQL_DB", "it_ticket")
	return Config{
		HTTPAddr:       httpAddr,
		MySQLDSN:       fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, port, db),
		JWTSecret:      jwtSecret,
		BootstrapEmail: bootEmail,
		BootstrapPass:  bootPass,
	}
}

// LogDSN 打日志用：把密码换成 ***。
func (c Config) LogDSN() string {
	dsn := c.MySQLDSN
	at := strings.Index(dsn, "@")
	colon := strings.Index(dsn, ":")
	if colon == -1 || at == -1 || colon > at {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
