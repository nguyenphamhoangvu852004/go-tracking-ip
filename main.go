package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type GeoInfo struct {
	Query      string `json:"query"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	Org        string `json:"org"`
	Timezone   string `json:"timezone"`
}

// Gọi API lấy thông tin địa lý theo IP
func GetGeoInfo(ip string) (*GeoInfo, error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var geo GeoInfo
	err = json.Unmarshal(body, &geo)
	if err != nil {
		return nil, err
	}
	return &geo, nil
}

// Hàm lấy IP thật sự từ X-Forwarded-For hoặc RemoteAddr
func GetRealIP(c *gin.Context) string {
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// Có thể chứa nhiều IP, lấy IP đầu tiên
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// Middleware ghi log và lưu IP vào context
func IPTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP := GetRealIP(c)
		c.Set("client_ip", realIP)

		geo, err := GetGeoInfo(realIP)
		if err != nil {
			log.Printf("📡 IP: %s - %s %s (Không xác định vị trí)", realIP, c.Request.Method, c.Request.URL.Path)
		} else {
			log.Printf("📡 IP: %s - %s, %s, %s | ISP: %s - %s %s",
				geo.Query, geo.City, geo.RegionName, geo.Country,
				geo.ISP, c.Request.Method, c.Request.URL.Path)
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()
	// Có thể dùng r.SetTrustedProxies nếu deploy thật, ở đây không cần
	r.Use(IPTrackingMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		ip, _ := c.Get("client_ip")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"ip":      ip,
		})
	})

	r.Run("0.0.0.0:8080")
}
