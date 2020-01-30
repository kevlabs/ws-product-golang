package middleware

import (
	"net"
	"net/http"
)

func getIPAddress(r *http.Request) (string, error) {
	var address string
	// check forwarded header in case user behind proxy
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		address = forwarded
	} else {
		address = r.RemoteAddr
	}

	IPAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	return IPAddress, nil
}
