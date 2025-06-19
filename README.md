# 🛡️ SSH Honeypot - Real-time SSH Brute-force Monitoring & Statistics System  
# 🛡️ SSH 蜜罐 - 实时 SSH 爆破监控与统计系统

A lightweight SSH honeypot system for real-time recording and visualization of SSH brute-force attacks from the internet.  
这是一个轻量级的 SSH 蜜罐系统，用于实时记录并可视化展示来自互联网的 SSH 爆破攻击行为。

With this tool, you can clearly see brute-force attempts against your system, including source IPs, usernames, and passwords.  
通过本工具，你可以清晰了解系统正在受到的爆破攻击情况，包括攻击源、使用的用户名和密码等关键信息。

---

## 📌 Features  
## 📌 功能特性

- Real-time logging of SSH login attempts  
  实时记录 SSH 登录尝试

- Automatically extract and count:  
  自动提取并统计以下信息：
  - Attack time 攻击时间  
  - Source IP 攻击来源 IP  
  - Source ASN 攻击来源 ASN  
  - Username & password 使用的用户名与密码

- Web UI with:
  提供 Web 页面展示：
  - Top IPs 高频 IP 分布  
  - Top ASNs 高频 ASN 分布  
  - Top Passwords 高频密码使用情况  

---

## 🚀 Quick Start (docker-compose)  
## 🚀 快速启动（docker-compose）

```yaml
version: '3.8'

services:
  ssh_honeypot:
    image: yessure/ssh_honeypot:latest
    container_name: ssh_honeypot
    restart: unless-stopped
    environment:
      MYSQL_DSN: "root:123456@tcp(172.17.0.1:3306)/honeypot?parseTime=true"
      REDIS_DSN: "redis://172.17.0.1:6379"  # optional; with password: redis://:mypassword@172.17.0.1:6379
    ports:
      - "22:22"                  # SSH honeypot port
      - "8000:8000"              # Web UI
    network_mode: bridge
```

> Adjust `MYSQL_DSN` and `REDIS_DSN` according to your environment.  
> 请根据实际环境调整 `MYSQL_DSN` 和 `REDIS_DSN`。

---

## 🧱 MySQL Table Initialization  
## 🧱 MySQL 表结构初始化

Please create the following tables in your database before running the container:  
请预先在数据库中创建如下表结构：

```sql
CREATE TABLE `asncount` (
  `asn` varchar(100) NOT NULL,
  `count` int(11) DEFAULT 0,
  PRIMARY KEY (`asn`),
  KEY `idx_count` (`count`)
);

CREATE TABLE `ipcount` (
  `ip` varchar(100) NOT NULL,
  `count` int(11) DEFAULT 0,
  PRIMARY KEY (`ip`),
  KEY `idx_count` (`count`)
);

CREATE TABLE `login_attempts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `ip` varchar(64) NOT NULL,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `attempt_time` bigint(20) NOT NULL,
  `asn` varchar(10) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `asn` (`asn`),
  KEY `idx_password` (`password`)
);

CREATE TABLE `passwordcount` (
  `password` varchar(100) NOT NULL,
  `count` int(11) DEFAULT 0,
  PRIMARY KEY (`password`),
  KEY `idx_count` (`count`)
);
```

---

## 📊 Web Interface  
## 📊 Web 页面展示

Visit `http://ip:8000` to view real-time statistics.  
访问 `http://ip:8000` 即可查看实时统计页面。

It is recommended to proxy this page behind a CDN or frontend server in production.  
实际使用中，建议使用前端 CDN/反代至本页。

To reduce access pressure, all displayed data is cached in Redis for 30 seconds.  
为降低并发访问压力，所有展示数据在 Redis 中设置了 30 秒缓存。

👉 Online Demo: [https://lostshit.com/api/login-attempts](https://lostshit.com/api/login-attempts)  

---

## 💡 Deployment Suggestion  
## 💡 建议部署方式

Deploy the container on a public server, mapping port 22 to attract real SSH attack traffic for analysis.  
建议将本容器部署于公网主机，通过端口映射将 22 端口暴露，以吸引真实 SSH 攻击流量用于分析。

