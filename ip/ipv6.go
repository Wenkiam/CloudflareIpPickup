package ip

import (
	"math/rand"
	"net"
)

type Ipv6 [16]byte
type Ipv6Range struct {
	network *Ipv6
	ipNet   *net.IPNet
}

func (ipv6 Ipv6) IsIpv4() bool {
	return false
}

func (ipv6 Ipv6) String() string {
	ip := make(net.IP, len(ipv6))
	copy(ip, ipv6[:])
	return ip.String()
}

func ParseIpv6(ip string) *Ipv6 {
	ipTemp := net.ParseIP(ip)
	ipv6 := new(Ipv6)
	copy(ipv6[:], ipTemp)
	return ipv6
}
func parseIpv6CIDR(cidr string) (*Ipv6Range, error) {
	ipv6, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ip := new(Ipv6)
	copy(ip[:], ipv6)
	return &Ipv6Range{ip, ipNet}, nil
}

func (ipv6Range Ipv6Range) pickup() []*Ip {
	ips := make([]*Ip, 0)
	ip := *ipv6Range.network
	for ipv6Range.contains(&ip) {
		if ipv6Range.ipNet.Mask[15] != 255 {
			ip[14] = byte(rand.Intn(255))
			ip[15] = byte(rand.Intn(255))
		}
		var ipToAppend Ip = ip
		ips = append(ips, &ipToAppend)
		for i := 13; i >= 0; i-- {
			tempIP := ip[i]
			ip[i] += byte(rand.Intn(255))
			if ip[i] >= tempIP {
				break
			}
		}
	}
	return ips
}

func (ipv6Range Ipv6Range) contains(ipv6 *Ipv6) bool {
	var ip = make(net.IP, 16)
	copy(ip, ipv6[:])
	return ipv6Range.ipNet.Contains(ip)
}
