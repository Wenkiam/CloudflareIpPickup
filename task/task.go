package task

import (
	"CloudflareIpPickup/ip"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"log"
	"sync"
	"time"
)

var (
	Port      = 443
	MaxDelay  = time.Second
	PingCount = 5
	TestCount = 10
)

type Task struct {
	ips       []*ip.Ipv4
	wg        *sync.WaitGroup
	mutex     *sync.Mutex
	pb        *pb.ProgressBar
	control   chan bool
	resultSet taskResultSet
}

type taskResult struct {
	ip      *ip.Ipv4
	rtt     time.Duration
	dlSpeed float64
}

type taskResultSet []*taskResult

func New() *Task {
	ips, err := ip.LoadIps()
	if err != nil {
		log.Fatal("load ips failed.reason:" + err.Error())
	}
	return &Task{
		ips,
		&sync.WaitGroup{},
		&sync.Mutex{},
		pb.StartNew(len(ips)),
		make(chan bool, 500),
		make(taskResultSet, 0),
	}
}

func (task *Task) Run() {
	rttTest(task)
	if !DownloadDisabled {
		downloadTest(task)
	}
	printResult(&task.resultSet)
	if Dns {
		setDnsRecords(&task.resultSet)
	}
}

func (result taskResultSet) Len() int {
	return len(result)
}
func (result taskResultSet) Less(i, j int) bool {
	iVal := float64(result[i].rtt) - result[i].dlSpeed
	jVal := float64(result[j].rtt) - result[j].dlSpeed
	return iVal < jVal
}
func (result taskResultSet) Swap(i, j int) {
	tmp := result[i]
	result[i] = result[j]
	result[j] = tmp
}

func printResult(resultSet *taskResultSet) {
	for index, result := range *resultSet {
		fmt.Printf("%-16s  %-16s %s\n", result.ip, result.rtt, speedToStr(result.dlSpeed))
		if index >= TestCount {
			break
		}
	}
}

func speedToStr(speed float64) string {
	const kb float64 = 1024
	const mb = kb * 1024
	const gb = mb * 1024
	if speed > gb {
		return fmt.Sprintf("%.2fGB/s", speed/gb)
	}
	if speed > mb {
		return fmt.Sprintf("%.2fMB/s", speed/mb)
	}
	if speed > kb {
		return fmt.Sprintf("%.2fKB/s", speed/kb)
	}
	return fmt.Sprintf("%.2fB/s", speed)
}
