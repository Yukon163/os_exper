package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var timeNow int
var ch = make(chan int)

type status int

const (
	Wait status = iota
	Run
	Finish
)

func (s status) String() string {
	switch s {
	case Wait:
		return "Wait"
	case Run:
		return "Run"
	case Finish:
		return "Finish"
	default:
		return "Unknown"
	}
}

//func GenerateRandInt(n int) int {
//	rand.New(rand.NewSource(time.Now().UnixNano()))
//	return rand.Intn(n)
//}

type PCB struct {
	name       string
	prior      int // 优先级 优先数？
	arriveTime int
	execTime   int64
	usedTime   int64
	PCBStatus  status
}

type processesQueue struct {
	Queue    []*PCB
	maxTime  int
	usedTime int
}

func showPCB(pcb PCB) {
	fmt.Printf("Process Name: %s, PCB Status: %s, PCB Prior: %d\n", pcb.name, pcb.PCBStatus, pcb.prior+1)
}

type CpuProcessScheduler struct {
	PCBQueues []*processesQueue
	cpuStatus status
}

func getCountFromPrior(prior int) int {
	return prior
}
func getPCBQueue(cpu *CpuProcessScheduler, QueueCount int) *[]*PCB {
	return &cpu.PCBQueues[QueueCount].Queue
}

func downgradePCBQueue(pcb *PCB) {
	if pcb.prior == 0 {
		return
	}
	fmt.Printf("%s's prior: %d->%d\n", pcb.name, pcb.prior+1, pcb.prior)
	pcb.prior -= 1
}

func InQueue(cpu *CpuProcessScheduler, pcb *PCB) {
	QueueCount := getCountFromPrior(pcb.prior) // 根据优先级获取队列
	PCBQueue := getPCBQueue(cpu, QueueCount)
	*PCBQueue = append(*PCBQueue, pcb)
	if cpu.cpuStatus == Wait {
		fmt.Printf("%s come, CPU Status: Waiting...\n", pcb.name)
		ch <- 1
	} else if cpu.cpuStatus == Run {
		fmt.Printf("%s come, CPU Status: Running...\n", pcb.name)
	} else {
		fmt.Printf("wt? cpu seems to be wrong, just check it's status\n")
		ch <- -1
	}
}

func OutQueue(cpu *CpuProcessScheduler, pcb *PCB) {
	for j := len(cpu.PCBQueues) - 1; j >= 0; j-- {
		PCBQueue := cpu.PCBQueues[j].Queue
		for i := len(PCBQueue) - 1; i >= 0; i-- {
			process := PCBQueue[i]
			if process.name == pcb.name {
				cpu.PCBQueues[j].Queue = append(PCBQueue[:i], PCBQueue[i+1:]...)
				return
			}
		}
	}
}

func stimulateCpuExecTime(pcb *PCB, PCBQueue *processesQueue) {
	time.Sleep(20 * time.Millisecond) // 模拟cpu执行进程花费的时间片
	pcb.usedTime += 1
	timeNow += 1
	PCBQueue.usedTime += 1
}

func changeCpuStatus(cpu *CpuProcessScheduler, statusToChange status) {
	statusNow := cpu.cpuStatus
	cpu.cpuStatus = statusToChange
	fmt.Printf("\nCPU Status: %s->%s \n", statusNow, statusToChange)
}

func showCpuProcessQueue(cpu *CpuProcessScheduler) {
	fmt.Printf("Now, the CPU's PCB queues are:\n")
	for i, PCBQueue := range cpu.PCBQueues {
		if len(PCBQueue.Queue) == 0 {
			fmt.Printf("PCBQueue%d is Null\n", i+1)
		} else {
			fmt.Printf("PCBQueue%d:\nused time: %d\n", i+1, PCBQueue.usedTime)
			for j, pcb := range PCBQueue.Queue {
				fmt.Printf("%d: ", j+1)
				showPCB(*pcb)
			}
		}
	}
}

func getNextHandleProcess(cpu *CpuProcessScheduler) (*PCB, int, bool) {
	var nextHandleProcess *PCB = nil
	var isExistNext = false
	var nextQueueCount = -1
	for i := len(cpu.PCBQueues) - 1; i >= 0; i-- {
		PCBQueue := cpu.PCBQueues[i]
		if len(PCBQueue.Queue) == 0 {
			fmt.Printf("Now, the CPU's PCBQueue%d is Null\nLooking for the nextQueue...\n", i+1)
			continue
		} else {
			if PCBQueue.usedTime != 0 {
				return PCBQueue.Queue[0], i, true
			} else if nextHandleProcess == nil {
				nextHandleProcess = PCBQueue.Queue[0]
				isExistNext = true
				nextQueueCount = i
			}
		}
	}
	return nextHandleProcess, nextQueueCount, isExistNext
}

func CpuHandleProcess(cpu *CpuProcessScheduler) {
	var pcb *PCB
	var value bool
	var QueueCount int

	pcb, QueueCount, value = getNextHandleProcess(cpu)
	if !value {
		fmt.Printf("Cpu's PCBQueues are null...")
		return
	}
	changeCpuStatus(cpu, Run)
	proQueue := cpu.PCBQueues[QueueCount]

	arrivalTime := pcb.arriveTime
	handleTime := timeNow
	fmt.Printf("handling process: %s at %d, which is arrived at %d, enter to the processes_queue...\n", pcb.name, handleTime, arrivalTime)

	fmt.Printf("handling process: %s...\n", pcb.name)
	stimulateCpuExecTime(pcb, proQueue)
	OutQueue(cpu, pcb)
	if pcb.usedTime < pcb.execTime {
		fmt.Printf("process %s still need to exec %d time, prior is %d\n", pcb.name, pcb.execTime-pcb.usedTime, pcb.prior+1)
		if proQueue.usedTime < proQueue.maxTime { // 一直运行到当前队列的MaxTime, 通过将其加入到当前队列队首实现
			proQueue.Queue = append([]*PCB{pcb}, proQueue.Queue...)
			//proQueue.usedTime += 1
		} else {
			downgradePCBQueue(pcb)
			proQueue.usedTime = 0
			InQueue(cpu, pcb)
		}
	} else {
		fmt.Printf("process:%s is finished\n", pcb.name)
		pcb.PCBStatus = Finish
		proQueue.usedTime = 0 // 任务完成同样需要将队列的使用时间清空
	}

	changeCpuStatus(cpu, Wait)
	ch <- 1
	showCpuProcessQueue(cpu)
	fmt.Printf("\n")
}

func processIn(cpu *CpuProcessScheduler, pcb *PCB) {
	pcb.arriveTime = timeNow
	QueueCount := getCountFromPrior(pcb.prior)
	for count, PCBQueue := range cpu.PCBQueues {
		if PCBQueue.usedTime != 0 && count < QueueCount {
			PCBQueue.usedTime = 0
			break
		}
	}
	InQueue(cpu, pcb)
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	cpu := CpuProcessScheduler{
		PCBQueues: []*processesQueue{ //与上次实验不同的是这里添加了显式初始化 maxTime 和 usedTime 的代码
			{Queue: []*PCB{}, maxTime: 3, usedTime: 0}, // 1
			{Queue: []*PCB{}, maxTime: 4, usedTime: 0}, // 2
			{Queue: []*PCB{}, maxTime: 5, usedTime: 0}, // 3
		},
		cpuStatus: Wait,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				cancelFunc()
				return
			case value := <-ch:
				if value == -1 {
					return
				}
				go CpuHandleProcess(&cpu)
			}
		}
	}()

	pcb1 := PCB{name: "P1", prior: 2, execTime: 4, usedTime: 0, PCBStatus: Wait}
	pcb2 := PCB{name: "P2", prior: 1, execTime: 5, usedTime: 0, PCBStatus: Wait}
	pcb3 := PCB{name: "P3", prior: 0, execTime: 4, usedTime: 0, PCBStatus: Wait}
	pcb4 := PCB{name: "P4", prior: 2, execTime: 11, usedTime: 0, PCBStatus: Wait}
	pcb5 := PCB{name: "P5", prior: 1, execTime: 3, usedTime: 0, PCBStatus: Wait}
	pcb6 := PCB{name: "P6", prior: 0, execTime: 6, usedTime: 0, PCBStatus: Wait}
	pcb7 := PCB{name: "P7", prior: 2, execTime: 10, usedTime: 0, PCBStatus: Wait}

	showCpuProcessQueue(&cpu)

	processIn(&cpu, &pcb1)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb2)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb3)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb4)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb5)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb6)
	time.Sleep(200 * time.Millisecond)
	processIn(&cpu, &pcb7)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancelFunc()
}
