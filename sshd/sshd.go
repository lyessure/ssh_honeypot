package sshd

import (
	"api/dao"
	"api/geo"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func StartSSHD() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成 RSA 密钥失败: %v", err)
	}
	private, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		log.Fatalf("转换 signer 失败: %v", err)
	}
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			remoteAddr := conn.RemoteAddr().String()
			ip := extractIP(remoteAddr)
			go logAttempt(ip, conn.User(), string(password))
			return nil, fmt.Errorf("permission denied")
		},
	}
	config.ServerVersion = "SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u6"
	config.AddHostKey(private)
	listener, err := net.Listen("tcp", ":22")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	log.Println("SSH 蜜罐已启动，监听端口 22")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}
		go handleConn(conn, config)
	}
}

func extractIP(addr string) string {
	// addr 格式可能是 ip:port，也可能纯 ip
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 如果出错，可能就是纯 ip，直接返回
		return addr
	}
	return ip
}

func handleConn(c net.Conn, config *ssh.ServerConfig) {
	defer c.Close()
	_, chans, reqs, err := ssh.NewServerConn(c, config)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)
	for ch := range chans {
		ch.Reject(ssh.Prohibited, "no shell here")
	}
}

func logAttempt(ip, username, password string) {
	now := time.Now().Unix()
	asn := resolveASN(ip)

	// 插入登录尝试记录
	if err := dao.GormDB.Exec(
		"INSERT INTO login_attempts (ip, username, password, asn, attempt_time) VALUES (?, ?, ?, ?, ?)",
		ip, username, password, asn, now,
	).Error; err != nil {
		log.Printf("写入登录尝试记录失败: %v", err)
	}

	// 更新密码计数
	if err := dao.GormDB.Exec(
		"INSERT INTO passwordcount (password, count) VALUES (?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		password,
	).Error; err != nil {
		log.Printf("更新密码计数失败: %v", err)
	}

	// 更新ASN计数
	if err := dao.GormDB.Exec(
		"INSERT INTO asncount (asn, count) VALUES (?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		asn,
	).Error; err != nil {
		log.Printf("更新ASN计数失败: %v", err)
	}

	// 更新IP计数
	if err := dao.GormDB.Exec(
		"INSERT INTO ipcount (ip, count) VALUES (?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		ip,
	).Error; err != nil {
		log.Printf("更新IP计数失败: %v", err)
	}
}

func resolveASN(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	info := geo.GetIPInfo(parsed.String())
	return info.PureASN
}
