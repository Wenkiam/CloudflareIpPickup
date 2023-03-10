package ip

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	IPv4len = 4
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Ipv4 [4]byte

type Ipv4Range struct {
	mask      *Ipv4
	network   *Ipv4
	broadcast *Ipv4
}

func (ip Ipv4) IsIpv4() bool {
	return true
}
func Parse(ip string) (*Ipv4, error) {
	ips := strings.Split(ip, ".")
	if len(ips) != IPv4len {
		return nil, errors.New("invalid ip format:" + ip)
	}
	p := new(Ipv4)
	for i := 0; i < IPv4len; i++ {
		b, err := strconv.Atoi(ips[i])
		if err != nil {
			return nil, err
		}
		p[i] = byte(b)
	}
	return p, nil
}

func parsIpv4CIDR(cidr string) (*Ipv4Range, error) {
	ss := strings.Split(cidr, "/")
	if len(ss) != 2 {
		return nil, errors.New("invalid format:" + cidr)
	}
	mask, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid mask:%s", ss[1]))
	}
	if mask > 32 || mask < 0 {
		return nil, errors.New(fmt.Sprintf("invalid mask:%d", mask))
	}
	ip, err := Parse(ss[0])
	if err != nil {
		return nil, errors.New("invalid format:" + cidr)
	}
	mask = -(1 << (32 - mask))
	ipRange := rangeOf(ip, intToIp(int32(mask)))
	return ipRange, nil
}

func rangeOf(ip, mask *Ipv4) *Ipv4Range {
	minIp := new(Ipv4)
	maxIp := new(Ipv4)
	for i := 0; i < 4; i++ {
		m := (*mask)[i]
		mm := ^m
		minIp[i] = (*ip)[i] & m
		maxIp[i] = mm | minIp[i]
	}
	return &Ipv4Range{mask, minIp, maxIp}
}

func (ipRange Ipv4Range) contains(ip *Ipv4) bool {
	min := ipRange.network.ipToInt32()
	max := ipRange.broadcast.ipToInt32()
	curr := ip.ipToInt32()
	return curr >= min && curr <= max
}

func (ip Ipv4) ipToInt32() int32 {

	return (int32(ip[0]) << 24) | (int32(ip[1]) << 16) | (int32(ip[2]) << 8) | int32(ip[3])
}

func intToIp(ip int32) *Ipv4 {
	p := new(Ipv4)
	var mask int32 = 0xff
	p[0] = byte((ip >> 24) & mask)
	p[1] = byte((ip >> 16) & mask)
	p[2] = byte((ip >> 8) & mask)
	p[3] = byte(ip & mask)
	return p
}

func (ip Ipv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ipRange Ipv4Range) String() string {
	return fmt.Sprintf("{mask:%s, Network:%s, broadcast:%s}", ipRange.mask, ipRange.network, ipRange.broadcast)
}

func (ipRange Ipv4Range) pickup() []*Ip {
	result := make([]*Ip, 0)
	network := (*ipRange.network)[3]
	randSeed := (*ipRange.broadcast)[3] - network
	var ip = *ipRange.network
	for ipRange.contains(&ip) {
		ip[3] = network + byte(rand.Intn(int(randSeed)))
		var ipToAppend Ip = ip
		result = append(result, &ipToAppend)
		ip[2]++
		if ip[2] == 0 {
			ip[1]++
			if ip[1] == 0 {
				ip[0]++
			}
		}
	}
	return result
}
