package geo

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

var (
	geoDBCity *geoip2.Reader
	geoDBASN  *geoip2.Reader
)

func InitGeoIP() error {
	var err error
	geoDBCity, err = geoip2.Open("data/GeoLite2-City.mmdb")
	if err != nil {
		log.Println("geoDBCity.Open error:", err)
		return err
	}
	geoDBASN, err = geoip2.Open("data/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Println("geoDBASN.Open error:", err)
		return err
	}
	return err
}

type IPInfo struct {
	Location string // 国家+城市
	ASN      string // AS编号+ISP
	PureASN  string
}

func GetIPInfo(ipStr string) IPInfo {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return IPInfo{}
	}

	info := IPInfo{}

	// 地理位置
	if geoDBCity != nil {
		if record, err := geoDBCity.City(ip); err == nil {
			country := record.Country.Names["zh-CN"]
			city := ""
			if len(record.Subdivisions) > 0 {
				city = record.Subdivisions[0].Names["zh-CN"]
			}
			if city == "" {
				city = record.City.Names["zh-CN"]
			}
			info.Location = strings.TrimSpace(country + " " + city)
		} else {
			log.Println("geoDBCity.City error:", err)
		}
	}

	// ASN
	if geoDBASN != nil {
		if record, err := geoDBASN.ASN(ip); err == nil {
			info.ASN = fmt.Sprintf("AS%d %s", record.AutonomousSystemNumber, record.AutonomousSystemOrganization)
			info.PureASN = fmt.Sprintf("AS%d", record.AutonomousSystemNumber)
		} else {
			log.Println("geoDBASN.ASN error:", err)
		}
	}

	return info
}
