package task

import (
	"CloudflareIpPickup/ip"
	"fmt"
	"net"
	"sort"
	"time"
)

func rttTest(task *Task) *Task {
	fmt.Printf("开始RTT测速,数量:%d,端口：%d，延迟上限：%v \n", len(task.ips), Port, MaxDelay)
	for _, ip := range task.ips {
		task.wg.Add(1)
		task.control <- false
		go task.ping(ip)
	}
	task.wg.Wait()
	task.pb.Finish()
	fmt.Println()
	sort.Sort(task.resultSet)
	return task
}

func min(nums ...int) int {
	res := nums[0]
	for _, num := range nums {
		if res > num {
			res = num
		}
	}
	return res
}

func (task *Task) ping(ip *ip.Ip) {
	defer task.wg.Done()
	defer task.pb.Add(1)
	pingResult := ping(ip)
	if pingResult.rtt == 0 {
		<-task.control
		return
	}
	task.mutex.Lock()
	defer task.mutex.Unlock()
	task.resultSet = append(task.resultSet, pingResult)
	<-task.control
}

func ping(ip *ip.Ip) *taskResult {
	result := taskResult{
		ip, 0, 0,
	}
	var address string
	if (*ip).IsIpv4() {
		address = fmt.Sprintf("%s:%d", *ip, Port)
	} else {
		address = fmt.Sprintf("[%s]:%d", *ip, Port)
	}
	total := time.Duration(0)
	okCount := 0
	for i := 0; i < PingCount; i++ {
		rtt, ok := tcpPing(&address)
		if ok {
			total += rtt
			okCount++
		}
	}
	if okCount != 0 {
		result.rtt = total / time.Duration(okCount)
	}
	return &result
}
func tcpPing(address *string) (time.Duration, bool) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", *address, MaxDelay)
	if err != nil {
		return 0, false
	}
	conn.Close()
	return time.Since(start), true
}
