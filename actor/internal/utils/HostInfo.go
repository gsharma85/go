package utils

import (
	"net"
)

func GetIpAddress(host string) string {
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
	    if ipv4 := addr.To4(); ipv4 != nil {
	        return ipv4.String()
	    }      
	}
	return ""
}