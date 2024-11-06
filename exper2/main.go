package main

import (
	"context"
	"fmt"
	"math/rand"
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

func GenerateRandInt(n int) int {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Intn(n)
}

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
	fmt.Printf("Process Name: %s, PCB Status: %s, PCB Prior: %d\n", pcb.name, pcb.PCBStatus, pcb.prior)
}

type CpuProcessScheduler struct {
	njmknu
	PCBQueues []*processesQueue
	cpuStatus status
}

func getCountFromPrior(prior int) int {
	return prior // 可拓展，根据不同大小的prior获取不同的count，这里为了方便只使用prior作为count
}
func getPCBQueue(cpu *CpuProcessScheduler, QueueCount int) []*PCB {
	return cpu.PCBQueues[QueueCount].Queue
}

func downgradePCBQueue(pcb *PCB) {
	pcb.prior -= 1
}

func InQueue(cpu *CpuProcessScheduler, pcb *PCB) {
	QueueCount := getCountFromPrior(pcb.prior) // 根据优先级获取队列
	PCBQueue := getPCBQueue(cpu, QueueCount)
	PCBQueue = append(PCBQueue, pcb)
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
	PCBQueue := getPCBQueue(cpu, getCountFromPrior(pcb.prior))
	for i, process := range PCBQueue {
		if process.name == pcb.name {
			PCBQueue = append(PCBQueue[:i], PCBQueue[i+1:]...)
			break
		}
	}
	ch <- 1
}

func stimulateCpuExecTime(pcb *PCB, PCBQueue *processesQueue) {
	time.Sleep(20 * time.Millisecond) // 模拟cpu执行进程花费的时间片
	pcb.usedTime += 1
	timeNow += 1
	PCBQueue.maxTime += 1
}

func changeCpuStatus(cpu *CpuProcessScheduler, statusToChange status) {
	fmt.Printf("\nCPU Status: %s->%s \n", cpu.cpuStatus, statusToChange)
	cpu.cpuStatus = statusToChange
}

func showCpuProcessQueue(cpu *CpuProcessScheduler) {
	fmt.Printf("Now, the CPU's PCB queues are:\n")
	for i, PCBQueue := range cpu.PCBQueues {
		if len(PCBQueue.Queue) == 0 {
			fmt.Printf("PCBQueue%d is Null\n", i+1)
		} else {
			fmt.Printf("PCBQueue%d:", i+1)
			for j, pcb := range PCBQueue.Queue {
				fmt.Printf("%d: , ", j)
				showPCB(*pcb)
			}
		}
	}
}

func getNextHandleProcess(cpu *CpuProcessScheduler) (*PCB, int, bool) {
	for i := len(cpu.PCBQueues) - 1; i >= 0; i-- {
		PCBQueue := cpu.PCBQueues[i]
		if len(PCBQueue.Queue) == 0 {
			fmt.Printf("Now, the CPU's PCBQueue%d is Null\nLooking for the nextQueue...\n", i+1)
			continue
		} else {
			return PCBQueue.Queue[0], i, true
		}
	}
	return nil, -1, false
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
	proQueue := cpu.PCBQueues[QueueCount]

	changeCpuStatus(cpu, Run)

	arrivalTime := pcb.arriveTime
	handleTime := timeNow
	fmt.Printf("handling process: %s at %d, which is arrived at %d, enter to the processes_queue...\n", pcb.name, handleTime, arrivalTime)

	fmt.Printf("handling process: %s...\n", pcb.name)
	stimulateCpuExecTime(pcb, proQueue)

	if pcb.usedTime < pcb.execTime {
		fmt.Printf("process %s still need to exec %d time\n", pcb.name, pcb.execTime-pcb.usedTime)
		if proQueue.usedTime < proQueue.maxTime { // 一直运行到当前队列的MaxTime, 通过将其加入到当前队列队首实现
			proQueue.Queue = append([]*PCB{pcb}, proQueue.Queue...)
		} else {
			downgradePCBQueue(pcb)
			proQueue.usedTime = 0
		}
	} else {
		fmt.Printf("process:%s is finished\n", pcb.name)
		pcb.PCBStatus = Finish
		proQueue.usedTime = 0 // 任务完成同样需要将队列的使用时间清空
		//defer wg.Done()
	}
	changeCpuStatus(cpu, Wait)

	OutQueue(cpu, pcb)
	showCpuProcessQueue(cpu)
	fmt.Printf("\n")
}

func processIn(cpu *CpuProcessScheduler, pcb *PCB) {
	pcb.arriveTime = timeNow
	InQueue(cpu, pcb)
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	cpu := CpuProcessScheduler{
		PCBQueues: []*processesQueue{&processesQueue{}, &processesQueue{}, &processesQueue{}}, // 三个空队列指针
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
	pcb4 := PCB{name: "P4", prior: 2, execTime: 2, usedTime: 0, PCBStatus: Wait}
	pcb5 := PCB{name: "P5", prior: 1, execTime: 3, usedTime: 0, PCBStatus: Wait}
	pcb6 := PCB{name: "P6", prior: 0, execTime: 3, usedTime: 0, PCBStatus: Wait}
	pcb7 := PCB{name: "P7", prior: 2, execTime: 3, usedTime: 0, PCBStatus: Wait}
	pcb8 := PCB{name: "P8", prior: 1, execTime: 3, usedTime: 0, PCBStatus: Wait}
	pcb9 := PCB{name: "P9", prior: 0, execTime: 3, usedTime: 0, PCBStatus: Wait}

	showPCB(pcb1)
	showPCB(pcb2)
	showPCB(pcb3)
	showPCB(pcb4)
	showPCB(pcb5)
	showPCB(pcb6)
	showPCB(pcb7)
	showPCB(pcb8)
	showPCB(pcb9)

	showCpuProcessQueue(&cpu)

	processIn(&cpu, &pcb1)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb2)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb3)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb4)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb5)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb6)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb7)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb8)
	time.Sleep(20 * time.Millisecond)
	processIn(&cpu, &pcb9)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancelFunc()
}
