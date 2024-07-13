package cyi

import (
	"net/http"
	"strings"
)
func getIp(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			ips := strings.Split(forwardedFor, ",")
			ip = ips[0]
		}
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}
