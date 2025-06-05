package main

import (
	"api/dao"
	"api/geo"
	"api/sshd"
	"api/sshmodule"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	geo.InitGeoIP()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("未指定数据库连接串，请通过 MYSQL_DSN 环境变量设置")
	}

	redisDSN := os.Getenv("REDIS_DSN")

	time.Local, _ = time.LoadLocation("Asia/Shanghai")
	dao.InitDB(dsn)

	if redisDSN != "" {
		dao.InitRedis(redisDSN)
	} else {
		log.Println("未配置 Redis，将直接访问数据库")
	}

	go sshd.StartSSHD()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/api/login-attempts", sshmodule.GetLatestLoginAttempts)
	r.GET("/api/ssh-toppasswords", sshmodule.GetTopPasswords)
	r.GET("/api/ssh-topasn", sshmodule.GetTopASN)
	r.GET("/api/ssh-topips", sshmodule.GetTopIPs)

	log.Println("服务运行在 http://0.0.0.0:8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
