package task

import (
	"CloudflareIpPickup/dns"
	"CloudflareIpPickup/util"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	DnsConfig = "cf_dns.conf"
	Dns       = false
)

func setDnsRecords(results *taskResultSet) {
	config := initConfig(DnsConfig)
	dns.AuthEmail = config["email"]
	if dns.AuthEmail == "" {
		dns.AuthEmail = util.ReadFromCmdLine("email")
	}
	dns.AuthKey = config["key"]
	if dns.AuthKey == "" {
		dns.AuthKey = util.ReadFromCmdLine("key")
	}
	dns.ZoneId = config["zoneId"]
	if dns.ZoneId == "" {
		dns.ZoneId = util.ReadFromCmdLine("zoneId")
	}
	domains := config["domains"]
	if domains == "" {
		domains = util.ReadFromCmdLine("域名，多个域名使用空格分开")
	}
	for index, domain := range strings.Fields(domains) {
		if index == len(*results) {
			break
		}
		dns.SetDnsRecord((*results)[index].ip.String(), strings.TrimSpace(domain))
	}
}
func initConfig(path string) map[string]string {
	config := make(map[string]string)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		fmt.Println("parse config file failed." + err.Error())
		return config
	}

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("parse config file failed." + err.Error())
			return config
		}
		s := strings.TrimSpace(string(b))
		if strings.Index(s, "#") == 0 {
			continue
		}

		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config
}
