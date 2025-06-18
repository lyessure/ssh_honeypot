package sshmodule

import (
	"api/dao"
	"api/geo"
	"api/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getLatestLoginAttemptsData() (interface{}, error) {
	type Attempt struct {
		IP          string
		Username    string
		Password    string
		AttemptTime string
		Location    string
	}

	var attempts []Attempt
	var loginAttempts []struct {
		IP          string `gorm:"column:ip"`
		Username    string `gorm:"column:username"`
		Password    string `gorm:"column:password"`
		AttemptTime int64  `gorm:"column:attempt_time"`
	}

	result := dao.GormDB.Model(&model.LoginAttempt{}).
		Select("ip, username, password, attempt_time").
		Order("attempt_time DESC").
		Limit(50).
		Find(&loginAttempts)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, la := range loginAttempts {
		var a Attempt
		a.IP = la.IP
		a.Username = la.Username
		a.Password = la.Password
		if strings.Contains(a.IP, ":") {
			a.IP = strings.Split(a.IP, ":")[0]
		}
		info := geo.GetIPInfo(a.IP)
		location := info.Location + "/" + info.ASN
		if len(location) > 45 {
			location = location[:42] + "…"
		}
		a.Location = location
		a.AttemptTime = time.Unix(la.AttemptTime, 0).In(time.Local).Format("2006-01-02 15:04:05")
		attempts = append(attempts, a)
	}

	var total int64
	if err := dao.GormDB.Model(&model.LoginAttempt{}).
		Select("COUNT(id)").
		Scan(&total).Error; err != nil {
		return nil, err
	}

	return struct {
		Attempts   []Attempt
		TotalCount int
	}{
		Attempts:   attempts,
		TotalCount: int(total),
	}, nil
}

func getTopPasswordsData() (interface{}, error) {
	type TopPassword struct {
		Password string
		Count    int
		Rank     int
	}
	var passwordCounts []struct {
		Password string `gorm:"column:password"`
		Count    int    `gorm:"column:count"`
	}
	if err := dao.GormDB.Model(&model.PasswordCount{}).
		Select("password, count").
		Order("count DESC").
		Limit(50).
		Find(&passwordCounts).Error; err != nil {
		return nil, err
	}
	var topPasswords []TopPassword
	for i, pc := range passwordCounts {
		topPasswords = append(topPasswords, TopPassword{
			Password: pc.Password,
			Count:    pc.Count,
			Rank:     i + 1,
		})
	}
	return struct {
		TopPasswords []TopPassword
	}{
		TopPasswords: topPasswords,
	}, nil
}

func getTopASNData() (interface{}, error) {
	type ASNStat struct {
		ASN   string
		Count int
		Rank  int
	}
	var asnCounts []struct {
		ASN   string `gorm:"column:asn"`
		Count int    `gorm:"column:count"`
	}
	if err := dao.GormDB.Model(&model.ASNCount{}).
		Select("asn, count").
		Order("count DESC").
		Limit(50).
		Find(&asnCounts).Error; err != nil {
		return nil, err
	}
	var list []ASNStat
	for i, ac := range asnCounts {
		asn := ac.ASN
		if asn == "" {
			asn = "未知"
		}
		var ip string
		dao.GormDB.Model(&model.LoginAttempt{}).
			Select("ip").
			Where("asn = ?", asn).
			Limit(1).
			Scan(&ip)
		var ipInfo geo.IPInfo
		if ip != "" {
			ipInfo = geo.GetIPInfo(ip)
		}
		location := ipInfo.ASN
		if len(location) > 45 {
			location = location[:42] + "…"
		}
		list = append(list, ASNStat{
			ASN:   location,
			Count: ac.Count,
			Rank:  i + 1,
		})
	}
	return struct {
		ASNList []ASNStat
	}{
		ASNList: list,
	}, nil
}

func getTopIPsData() (interface{}, error) {
	type IPStat struct {
		IP       string
		Count    int
		Rank     int
		Location string
	}
	var ipCounts []struct {
		IP    string `gorm:"column:ip"`
		Count int    `gorm:"column:count"`
	}
	if err := dao.GormDB.Model(&model.IPCount{}).
		Select("ip, count").
		Order("count DESC").
		Limit(50).
		Find(&ipCounts).Error; err != nil {
		return nil, err
	}
	var list []IPStat
	for i, ic := range ipCounts {
		ip := ic.IP
		if ip == "" {
			ip = "未知"
		}
		ipInfo := geo.GetIPInfo(ip)
		location := ipInfo.Location + "/" + ipInfo.ASN
		if len(location) > 45 {
			location = location[:42] + "…"
		}
		list = append(list, IPStat{
			IP:       ip,
			Count:    ic.Count,
			Rank:     i + 1,
			Location: location,
		})
	}
	return struct {
		IPList []IPStat
	}{
		IPList: list,
	}, nil
}

func GetLatestLoginAttempts(c *gin.Context) {
	type Attempt struct {
		IP          string
		Username    string
		Password    string
		AttemptTime string
		Location    string
	}
	var data struct {
		Attempts   []Attempt
		TotalCount int
	}
	err := dao.LoadDataWithLock("latest_login_attempts", &data, getLatestLoginAttemptsData, 30)
	if err != nil {
		c.String(http.StatusInternalServerError, "系统错误")
		return
	}
	c.HTML(http.StatusOK, "ssh_login.html", data)
}

func GetTopPasswords(c *gin.Context) {
	type TopPassword struct {
		Password string
		Count    int
		Rank     int
	}
	var data struct {
		TopPasswords []TopPassword
	}
	err := dao.LoadDataWithLock("top_password_stats", &data, getTopPasswordsData, 30)
	if err != nil {
		c.String(http.StatusInternalServerError, "系统错误")
		return
	}
	c.HTML(http.StatusOK, "ssh_toppasswords.html", data)
}

func GetTopASN(c *gin.Context) {
	type ASNStat struct {
		ASN   string
		Count int
		Rank  int
	}
	var data struct {
		ASNList []ASNStat
	}
	err := dao.LoadDataWithLock("top_asn_stats", &data, getTopASNData, 30)
	if err != nil {
		c.String(http.StatusInternalServerError, "系统错误")
		return
	}
	c.HTML(http.StatusOK, "ssh_topasn.html", data)
}

func GetTopIPs(c *gin.Context) {
	type IPStat struct {
		IP       string
		Count    int
		Rank     int
		Location string
	}
	var data struct {
		IPList []IPStat
	}
	err := dao.LoadDataWithLock("top_ip_stats", &data, getTopIPsData, 30)
	if err != nil {
		c.String(http.StatusInternalServerError, "系统错误")
		return
	}
	c.HTML(http.StatusOK, "ssh_topips.html", data)
}
