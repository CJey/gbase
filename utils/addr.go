package utils

import (
	"net"
	"strconv"
	"strings"
)

// AddrCompletion convert given raw string to standardized address,
// if the ip/port part of raw missing, will use ip and port to complete it.
// if the given ip is nil, treat as 0.0.0.0
// ""      => ip:port
// 80      => ip:80
// 1.1.1.1 => 1.1.1.1:port
func AddrCompletion(raw string, ip net.IP, port uint16) string {
	var (
		istr string = "0.0.0.0"
		pstr string = strconv.FormatUint(uint64(port), 10)
	)
	if ip.To4() != nil || ip.To16() != nil {
		istr = ip.String()
	}
	raw = strings.TrimSpace(raw)

	if raw == "" {
		// empty
		raw = istr + ":" + pstr
	} else if rawport, err := strconv.ParseUint(raw, 10, 16); err == nil {
		// raw port
		raw = istr + ":" + strconv.FormatUint(uint64(rawport), 10)
	} else if rawip := net.ParseIP(raw); rawip != nil {
		// raw ip
		raw = rawip.String() + ":" + pstr
	}
	return raw
}

// ResolveTCPAddr use AddrCompletion to resolve the raw as a tcp address
func ResolveTCPAddr(raw string, ip net.IP, port uint16) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", AddrCompletion(raw, ip, port))
}

// ResolveUDPAddr use AddrCompletion to resolve the raw as a udp address
func ResolveUDPAddr(raw string, ip net.IP, port uint16) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", AddrCompletion(raw, ip, port))
}
