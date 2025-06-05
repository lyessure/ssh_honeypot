package dao

import (
	"database/sql"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var GormDB *gorm.DB

func InitDB(dsn string) {
	var err error

	Db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err = Db.Ping(); err != nil {
		log.Fatalf("数据库不可用: %v", err)
	}

	Db.SetMaxOpenConns(50)           // 设置最大打开连接数
	Db.SetMaxIdleConns(5)            // 设置最大空闲连接数
	Db.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期

	GormDB, err = gorm.Open(mysql.New(mysql.Config{
		Conn: Db, // 重用现有的数据库连接
	}), &gorm.Config{})

	log.Println("数据库连接成功")
}
