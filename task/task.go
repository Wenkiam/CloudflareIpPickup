package task

import (
	"CloudflareIpPickup/ip"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"log"
	"os"
	"sync"
	"time"
)

var (
	Port       = 443
	MaxDelay   = time.Second
	PingCount  = 5
	TestCount  = 10
	OutPutFile = "result.txt"
)

type Task struct {
	ips       []*ip.Ip
	wg        *sync.WaitGroup
	mutex     *sync.Mutex
	pb        *pb.ProgressBar
	control   chan bool
	resultSet taskResultSet
}

type taskResult struct {
	ip      *ip.Ip
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
	contents := printResult(&task.resultSet)
	if Dns {
		setDnsRecords(&task.resultSet)
	}
	file, err := os.OpenFile(OutPutFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("open out put file failed." + err.Error())
	}
	defer file.Close()
	fmt.Println("准备将结果写入文件：" + OutPutFile)
	for _, content := range contents {
		file.WriteString(content)
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

func printResult(resultSet *taskResultSet) []string {
	content := make([]string, 0)
	for index, result := range *resultSet {
		line := fmt.Sprintf("%-16s  %-16s %s\n", *result.ip, result.rtt, speedToStr(result.dlSpeed))
		fmt.Print(line)
		content = append(content, line)
		if index >= TestCount {
			break
		}
	}
	return content
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
