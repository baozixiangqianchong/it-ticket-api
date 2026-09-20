package store

import (
	"database/sql"
	"time"

	// 匿名导入：只跑驱动的 init，把名字 "mysql" 登记给 database/sql。
	_ "github.com/go-sql-driver/mysql"
)

// OpenMySQL 连上 MySQL 并设好连接池。
// 返回的 *sql.DB 是连接池，整个进程共用。这里不建表。
func OpenMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
