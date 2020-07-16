package gossip

import (
	"net"
	"strconv"
	"strings"
)

func UniqTCPAddr(one string, all []string) (*net.TCPAddr, []*net.TCPAddr, error) {
	var addr, err = net.ResolveTCPAddr("tcp", one)
	if err != nil {
		return nil, nil, err
	}
	var addrs = make([]*net.TCPAddr, 0)
	var uniq = make(map[string]bool)
	for _, a := range all {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		var a, e = net.ResolveTCPAddr("tcp", a)
		if e != nil {
			return nil, nil, e
		}
		if a.String() == addr.String() {
			continue
		}
		if uniq[a.String()] == false {
			addrs = append(addrs, a)
			uniq[a.String()] = true
		}
	}
	return addr, addrs, nil
}

func ParseCSVAddrs(csv string, port uint16) ([]string, error) {
	uniq := make(map[string]bool)
	addrs := make([]string, 0)
	for _, val := range strings.Split(csv, ",") {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		rawip := net.ParseIP(val)
		if rawip.To4() != nil {
			val = rawip.String() + ":" + strconv.Itoa(int(port))
		}
		addr, err := net.ResolveTCPAddr("tcp", val)
		if err != nil {
			return nil, err
		}
		if uniq[addr.String()] == false {
			addrs = append(addrs, addr.String())
			uniq[addr.String()] = true
		}
	}
	return addrs, nil
}
