package main

import (
	"fmt"
	"math/rand"
	//"sync"
	"time"
)

var timeNow int

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
	rand.Seed(time.Now().UnixNano())
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

func showPCB(pcb PCB) {
	fmt.Printf("Process Name: %s, PCBStatus: %s\n", pcb.name, pcb.PCBStatus)
}

type CpuProcessScheduler struct {
	processesQueue []PCB
	cpuStatus      status
}

func InQueue(cpu *CpuProcessScheduler, pcb PCB) {
	//fmt.Printf("Process: %s move into queue", pcb.name)
	cpu.processesQueue = append(cpu.processesQueue, pcb)
}

func OutQueue(cpu *CpuProcessScheduler, pcb PCB) {
	for i, process := range cpu.processesQueue {
		if process.name == pcb.name {
			cpu.processesQueue = append(cpu.processesQueue[:i], cpu.processesQueue[i+1:]...)
			//fmt.Printf("Process: %s removed from queue\n", pcb.name)
			break
		}
	}
}

func showCpuProcessQueue(cpu *CpuProcessScheduler) {
	fmt.Printf("Now, the CPU's process_queue is:\n")
	if len(cpu.processesQueue) == 0 {
		for i, pcb := range cpu.processesQueue {
			fmt.Printf("%d: ", i)
			showPCB(pcb)
		}
	} else {

	}
}

func findHighestPriorProcess(cpu *CpuProcessScheduler) PCB {
	if len(cpu.processesQueue) == 0 {
		fmt.Printf("No processes queue\n")
	}

	highestPriorProcess := cpu.processesQueue[0]
	for _, process := range cpu.processesQueue {
		if process.prior > highestPriorProcess.prior {
			highestPriorProcess = process
		}
	}
	return highestPriorProcess
}

func showCPU(cpu *CpuProcessScheduler) {
	fmt.Printf("CPUStatus: %s\n", cpu.cpuStatus)
}

func stimulateCpuExecTime(pcb PCB) {
	time.Sleep(10) // 模拟cpu执行进程花费的时间片
	pcb.usedTime += 1
	timeNow += 1
}

func changeCpuStatus(cpu *CpuProcessScheduler, statusToChange status) {
	fmt.Printf("\nCPU Status: %s->%s \n", cpu.cpuStatus, statusToChange)
	cpu.cpuStatus = Run
}

func CpuHandleProcess(cpu *CpuProcessScheduler) {
	pcb := findHighestPriorProcess(cpu)
	OutQueue(cpu, pcb)

	arrivalTime := pcb.arriveTime
	handleTime := timeNow

	changeCpuStatus(cpu, Run)
	fmt.Printf("handling process: %s at %d, arriving at %d, enter to the processes_queue...\n", pcb.name, handleTime, arrivalTime)

	fmt.Printf("start process: %s...\n", pcb.name)
	stimulateCpuExecTime(pcb)
	if pcb.usedTime < pcb.execTime {
		fmt.Printf("process still need to exec %d time\n", pcb.execTime-pcb.usedTime)
		pcb.prior -= 1
	} else {
		fmt.Printf("process:%s is finished\n", pcb.name)
		pcb.PCBStatus = Finish
	}
	changeCpuStatus(cpu, Wait)
	showCpuProcessQueue(cpu)
}

func processIn(cpu *CpuProcessScheduler, pcb PCB) {
	pcb.arriveTime = timeNow
	InQueue(cpu, pcb)
	fmt.Printf("CPU Status: %s...\nProcess: %s enter to the processes_queue...\n", cpu.cpuStatus, pcb.name)

	if cpu.cpuStatus == Wait {
		//go CpuHandleProcess(cpu)
		CpuHandleProcess(cpu)
	} else if cpu.cpuStatus == Run {
		fmt.Printf("CPU Status: %s...\nprocess: %s enter to the processes_queue...\n", cpu.cpuStatus, pcb.name)
		showCpuProcessQueue(cpu)
	} else {
		fmt.Printf("wt? cpu seems to be wrong, just check it's status\n")
	}
}

func scheduleProcesses(cpu *CpuProcessScheduler, pcbs []PCB) {
	for _, pcb := range pcbs {
		go processIn(cpu, pcb)
	}
}

func main() {
	cpu := CpuProcessScheduler{
		processesQueue: nil,
		cpuStatus:      Wait,
	}
	pcb1 := PCB{
		name:       "P1",
		prior:      GenerateRandInt(100),
		arriveTime: 0,
		execTime:   10,
		usedTime:   0,
		PCBStatus:  Wait,
	}
	//showCPU(&cpu)
	//showPCB(pcb1)
	processIn(&cpu, pcb1)
}
