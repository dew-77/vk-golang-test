package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"
)

type ContainerStatus struct {
	IPAddress   string    `json:"ip_address"`
	PingStatus  bool      `json:"ping_status"`
	LastChecked time.Time `json:"last_checked"`
}

func ping(ip string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "80"), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func main() {
	ips := []string{"127.0.0.1", "192.168.0.1"}

	for {
		for _, ip := range ips {
			status := ContainerStatus{
				IPAddress:   ip,
				PingStatus:  ping(ip),
				LastChecked: time.Now(),
			}

			data, err := json.Marshal(status)
			if err != nil {
				log.Printf("Ошибка при сериализации данных для IP %s: %v", ip, err)
				continue
			}

			resp, err := http.Post("http://backend:8080/data", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Ошибка при отправке данных для IP %s: %v", ip, err)
				continue
			}
			resp.Body.Close()

			log.Printf("Данные для IP %s отправлены: статус=%v", ip, status.PingStatus)
		}

		time.Sleep(5 * time.Minute)
	}
}
