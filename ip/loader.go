package ip

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	CfIpv4Url = "https://www.cloudflare.com/ips-v4"
)

var (
	Ipv4Url = CfIpv4Url
	File    = "ip.txt"
)

func LoadIps() ([]*Ip, error) {
	ips, err := loadFromFile()
	if err == nil && len(ips) > 0 {
		return ips, nil
	}
	res, err := http.Get(Ipv4Url)
	if err != nil {
		log.Println("load ip list from dns failed,reason:" + err.Error())
		return nil, err
	}
	scanner := bufio.NewScanner(res.Body)
	return loadFromScanner(scanner), nil
}
func loadFromFile() ([]*Ip, error) {
	file, err := os.Open(File)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	return loadFromScanner(scanner), nil
}
func loadFromScanner(scanner *bufio.Scanner) []*Ip {
	ranges := make([]*ipRange, 0)
	for scanner.Scan() {
		text := scanner.Text()
		ip, err := parseCIDR(text)
		if err != nil {
			log.Println("parse ip failed, ip = " + text + ".reason:" + err.Error())
			continue
		}
		fmt.Println(text)
		ranges = append(ranges, ip)
	}
	ips := make([]*Ip, 0)
	for i := 0; i < len(ranges); i++ {
		ips = append(ips, (*ranges[i]).pickup()...)
	}
	return ips
}
func parseCIDR(cidr string) (*ipRange, error) {
	isIpv6 := strings.Contains(cidr, ":")
	cidr = validate(cidr, isIpv6)
	var res ipRange
	if isIpv6 {
		ipv6Range, err := parseIpv6CIDR(cidr)
		res = ipv6Range
		return &res, err
	} else {
		ipv4Range, err := parsIpv4CIDR(cidr)
		res = ipv4Range
		return &res, err
	}
}
func validate(ip string, isIpv6 bool) string {
	if strings.Contains(ip, "/") {
		return ip
	}
	if isIpv6 {
		return ip + "/128"
	}
	return ip + "/32"
}
