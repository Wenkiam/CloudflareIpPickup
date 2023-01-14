package ip

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

const (
	CfIpv4Url = "https://www.cloudflare.com/ips-v4"
)

var (
	Ipv4Url = CfIpv4Url
)

func LoadIps() ([]*Ipv4, error) {
	res, err := http.Get(Ipv4Url)
	if err != nil {
		log.Println("load ip list from dns failed,reason:" + err.Error())
		return nil, err
	}
	scanner := bufio.NewScanner(res.Body)
	ranges := make([]*Ipv4Range, 0)
	for scanner.Scan() {
		text := scanner.Text()
		ipWithMask := validate(text)
		ip, err := parsCIDR(ipWithMask)
		if err != nil {
			log.Println("parse ip failed, ip = " + text + ".reason:" + err.Error())
			continue
		}
		fmt.Println(text)
		ranges = append(ranges, ip)
	}
	ips := make([]*Ipv4, 0)
	for i := 0; i < len(ranges); i++ {
		ips = append(ips, ranges[i].pickup()...)
	}
	return ips, nil
}
func validate(ip string) string {
	var mask string
	if i := strings.IndexByte(ip, '/'); i < 0 {
		mask = "/32"
		ip += mask
	} else {
		mask = ip[i:]
	}
	return ip
}

func (ipRange *Ipv4Range) pickup() []*Ipv4 {
	result := make([]*Ipv4, 0)
	network := (*ipRange.network)[3]
	randSeed := (*ipRange.broadcast)[3] - network
	var ip = *ipRange.network
	for ipRange.contains(&ip) {
		ip[3] = network + byte(rand.Intn(int(randSeed)))
		ipToAppend := ip
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
