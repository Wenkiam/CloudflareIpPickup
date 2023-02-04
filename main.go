package main

import (
	"CloudflareIpPickup/ip"
	"CloudflareIpPickup/task"
	"flag"
	"time"
)

func init() {
	var maxDelay int
	flag.IntVar(&task.Port, "p", 443, "rtt测速端口")
	flag.IntVar(&task.PingCount, "pc", 5, "每个ip ping次数")
	flag.IntVar(&maxDelay, "t", 999, "ping超时时间")
	flag.IntVar(&task.TestCount, "n", 10, "挑选的ip数量")
	task.MaxDelay = time.Duration(maxDelay) * time.Millisecond
	flag.StringVar(&ip.Ipv4Url, "url", ip.CfIpv4Url, "ip列表下载地址")
	flag.BoolVar(&task.Dns, "dns", false, "测速完成后是否进行dns解析")
	flag.BoolVar(&task.DownloadDisabled, "dd", false, "是否禁用下载速度测试")
	flag.StringVar(&task.DnsConfig, "dp", "cf_dns.conf", "域名解析配置文件路径")
	flag.StringVar(&task.URL, "du", task.DefaultURL, "下载速度测试url")
	flag.StringVar(&task.OutPutFile, "rf", "result.txt", "测试结果文件路径")
	flag.StringVar(&ip.File, "f", "ip.txt", "要测试的ip文件路径")
	flag.Parse()

}
func main() {
	task.New().Run()
}
func usage() {
	flag.Usage()

}
