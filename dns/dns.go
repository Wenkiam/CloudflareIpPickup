package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	recordListUrl   = "https://api.cloudflare.com/client/v4/zones/%s/dns_records"
	recordUpdateUrl = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
	recordAddUrl    = "https://api.cloudflare.com/client/v4/zones/%s/dns_records"
)

var (
	ZoneId    = ""
	AuthEmail = ""
	AuthKey   = ""
)

type record struct {
	name    string
	id      string
	rType   string
	ip      string
	proxied bool
	ttl     int
}

func loadDnsRecords(zone string) map[string]*record {
	request, err := http.NewRequest("GET", fmt.Sprintf(recordListUrl, zone), nil)
	if err != nil {
		log.Fatal("create request failed." + err.Error())
		return nil
	}
	request.Header.Set("X-Auth-Email", AuthEmail)
	request.Header.Set("X-Auth-Key", AuthKey)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal("request for dns records failed." + err.Error())
		return nil
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	return parseResponse(&bytes)
}
func parseResponse(response *[]byte) map[string]*record {
	var data map[string]interface{}
	json.Unmarshal(*response, &data)
	success := data["success"].(bool)
	if !success {
		errors := data["errors"]
		log.Fatalln(errors)
		return nil
	}
	array := data["result"].([]interface{})
	records := make(map[string]*record)
	for _, result := range array {
		item := result.(map[string]interface{})
		r := record{
			item["name"].(string),
			item["id"].(string),
			item["type"].(string),
			item["content"].(string),
			item["proxied"].(bool),
			int(item["ttl"].(float64)),
		}
		records[r.name] = &r
	}
	return records
}

func SetDnsRecord(ip, domain string) {
	records := loadDnsRecords(ZoneId)
	r := records[domain]
	method := "PUT"
	url := fmt.Sprintf(recordAddUrl, ZoneId)
	if r == nil {
		method = "POST"
		r = &record{
			name:    domain,
			ip:      ip,
			ttl:     1,
			proxied: false,
			rType:   "A",
		}
	} else {
		url = fmt.Sprintf(recordUpdateUrl, ZoneId, r.id)
	}
	body := fmt.Sprintf("{\"type\":\"%s\",\"name\":\"%s\",\"content\":\"%s\",\"ttl\":%d,\"proxied\":%v}",
		r.rType, r.name, r.ip, r.ttl, r.proxied)
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	defer request.Body.Close()
	if err != nil {
		log.Fatalln("create request failed." + err.Error())
		return
	}
	request.Header.Set("X-Auth-Email", AuthEmail)
	request.Header.Set("X-Auth-Key", AuthKey)

	response, err := http.DefaultClient.Do(request)
	bytes, err := io.ReadAll(response.Body)
	result := make(map[string]interface{}, 0)
	json.Unmarshal(bytes, &result)
	success := result["success"].(bool)
	if success {
		fmt.Printf("set domain %s to %s success\n", domain, ip)
	} else {
		fmt.Printf("set domain %s to %s failed.%v\n", domain, ip, result["errors"])
	}
}
