package main

import (
	"fmt"
	"math/rand"
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
			fmt.Printf("Now, the CPU's PCBQueue%d is Null\n", i+1)
		} else {
			for j, pcb := range PCBQueue.Queue {
				fmt.Printf("the PCB queue%d:\n", i+1)
				fmt.Printf("%d: , ", j)
				showPCB(*pcb)
			}
		}
	}
}

func getNextHandleProcess(cpu *CpuProcessScheduler) (*PCB, int, bool) {
	for i, PCBQueue := range cpu.PCBQueues {
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

func main() {
	cpu := CpuProcessScheduler{
		PCBQueues: []*processesQueue{&processesQueue{}, &processesQueue{}, &processesQueue{}},
		cpuStatus: Wait, // 确保 Wait 是一个有效的 status 值
	}
	showCpuProcessQueue(&cpu)
}
