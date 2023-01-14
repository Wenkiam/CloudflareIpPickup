package task

import (
	"CloudflareIpPickup/ip"
	"context"
	"fmt"
	"github.com/VividCortex/ewma"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net"
	"net/http"
	"sort"
	"time"
)

const (
	bufferSize = 1024
	DefaultURL = "https://speed.cloudflare.com/__down?bytes=100000000"
)

var (
	URL              = DefaultURL
	DownloadDisabled = false
	Timeout          = time.Second * 10
)

func downloadTest(task *Task) {
	fmt.Println("开始下载速度测试")
	task.pb = pb.StartNew(min(TestCount, len(task.resultSet)))
	resultSet := make(taskResultSet, 0)
	for _, r := range task.resultSet {
		speed := downloadHandler(r.ip)
		if speed > 0 {
			r.dlSpeed = speed
			task.pb.Add(1)
			resultSet = append(resultSet, r)
		}
		if len(resultSet) >= TestCount {
			break
		}
	}
	task.pb.Finish()
	fmt.Println()
	sort.Sort(resultSet)
	task.resultSet = resultSet
}

func getDialContext(ip *ip.Ipv4) func(ctx context.Context, network, address string) (net.Conn, error) {
	var fakeSourceAddr string
	fakeSourceAddr = fmt.Sprintf("%s:%d", ip.String(), Port)
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, fakeSourceAddr)
	}
}

// return download Speed
func downloadHandler(ip *ip.Ipv4) float64 {
	client := &http.Client{
		Transport: &http.Transport{DialContext: getDialContext(ip)},
		Timeout:   Timeout,
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return 0.0
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")

	response, err := client.Do(req)
	if err != nil {
		return 0.0
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return 0.0
	}
	timeStart := time.Now()           // 开始时间（当前）
	timeEnd := timeStart.Add(Timeout) // 加上下载测速时间得到的结束时间

	contentLength := response.ContentLength // 文件大小
	buffer := make([]byte, bufferSize)

	var (
		contentRead     int64 = 0
		timeSlice             = Timeout / 100
		timeCounter           = 1
		lastContentRead int64 = 0
	)

	var nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
	e := ewma.NewMovingAverage()

	// 循环计算，如果文件下载完了（两者相等），则退出循环（终止测速）
	for contentLength != contentRead {
		currentTime := time.Now()
		if currentTime.After(nextTime) {
			timeCounter++
			nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
			e.Add(float64(contentRead - lastContentRead))
			lastContentRead = contentRead
		}
		// 如果超出下载测速时间，则退出循环（终止测速）
		if currentTime.After(timeEnd) {
			break
		}
		bufferRead, err := response.Body.Read(buffer)
		if err != nil {
			if err != io.EOF { // 文件下载完了，或因网络等问题导致链接中断，则退出循环（终止测速）
				break
			}
			e.Add(float64(contentRead-lastContentRead) / (float64(nextTime.Sub(currentTime)) / float64(timeSlice)))
		}
		contentRead += int64(bufferRead)
	}
	return e.Value() / (Timeout.Seconds() / 120)
}
