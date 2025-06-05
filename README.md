# 🛡️ SSH Honeypot - 实时 SSH 爆破监控与统计系统

这是一个轻量级的 SSH 蜜罐系统，用于实时记录并可视化展示来自互联网的 SSH 爆破攻击行为。  
通过本工具，你可以清晰了解系统正在受到的爆破攻击情况，包括攻击源、使用的用户名和密码等关键信息。

---

## 📌 功能特性

- 实时记录 SSH 登录尝试
- 自动提取并统计以下信息：
  - 攻击时间
  - 攻击来源 IP
  - 攻击来源 ASN
  - 使用的用户名与密码
- 提供 Web 页面展示：
  - 高频 IP 分布
  - 高频 ASN 分布
  - 高频密码使用情况

---

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
      REDIS_DSN: "redis://172.17.0.1:6379"  # 可不设定；若有密码：redis://:mypassword@172.17.0.1:6379
    ports:
      - "22:22"                  # SSH 蜜罐监听端口
      - "8000:8000"              # Web 统计页
    network_mode: bridge
```

> 请根据实际环境调整 `MYSQL_DSN` 和 `REDIS_DSN`。

---

## 🧱 MySQL 表结构初始化

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

## 📊 Web 页面展示

访问 `http://ip:8000` 即可查看实时统计页面。实际使用中，建议使用前端 CDN/反代至本页。  
为降低并发访问压力，所有展示数据在 Redis 中设置了 30 秒缓存。

👉 在线 Demo：[https://lostshit.com/api/login-attempts](https://lostshit.com/api/login-attempts)



---

## 💡 建议部署方式

建议将本容器部署于公网主机，通过端口映射将 22 端口暴露，以吸引真实 SSH 攻击流量用于分析。
