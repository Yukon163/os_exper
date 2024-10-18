package main

import (
	"fmt"
	"math/rand"
	"sync"

	//"sync"
	"time"
)

var timeNow int
var ch = make(chan int)
var ch1 = make(chan int)
var wg = sync.WaitGroup{}

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

func showPCB(pcb PCB) {
	fmt.Printf("Process Name: %s, PCB Status: %s, PCB Prior: %d\n", pcb.name, pcb.PCBStatus, pcb.prior)
}

type CpuProcessScheduler struct {
	processesQueue []PCB
	cpuStatus      status
}

func InQueue(cpu *CpuProcessScheduler, pcb *PCB) {
	//fmt.Printf("Process: %s move into queue", pcb.name)
	cpu.processesQueue = append(cpu.processesQueue, *pcb)
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
	for i, process := range cpu.processesQueue {
		if process.name == pcb.name {
			cpu.processesQueue = append(cpu.processesQueue[:i], cpu.processesQueue[i+1:]...)
			//fmt.Printf("Process: %s removed from queue\n", pcb.name)
			break
		}
	}
	ch <- 1
}

func showCpuProcessQueue(cpu *CpuProcessScheduler) {
	if len(cpu.processesQueue) != 0 {
		fmt.Printf("Now, the CPU's process_queue is:\n")
		for i, pcb := range cpu.processesQueue {
			fmt.Printf("%d: ", i)
			showPCB(pcb)
		}
		HighestPriorProcess, _ := findHighestPriorProcess(cpu)
		fmt.Printf("the HighestPriorProcess is %s\n", HighestPriorProcess.name)
	} else {
		fmt.Printf("Now, the CPU's process_queue is Null\n")
	}
}

func findHighestPriorProcess(cpu *CpuProcessScheduler) (*PCB, bool) {
	if len(cpu.processesQueue) == 0 {
		fmt.Printf("No processes queue\n")
		return &PCB{}, false
	} else {
		highestPriorProcess := cpu.processesQueue[0]
		for _, process := range cpu.processesQueue {
			if process.prior > highestPriorProcess.prior {
				highestPriorProcess = process
			}
		}
		return &highestPriorProcess, true
	}
}

func stimulateCpuExecTime(pcb *PCB) {
	time.Sleep(10 * time.Millisecond) // 模拟cpu执行进程花费的时间片
	pcb.usedTime += 1
	timeNow += 1
	//ch <- 1
}

func changeCpuStatus(cpu *CpuProcessScheduler, statusToChange status) {
	fmt.Printf("\nCPU Status: %s->%s \n", cpu.cpuStatus, statusToChange)
	cpu.cpuStatus = statusToChange
}

func CpuHandleProcess(cpu *CpuProcessScheduler) {
	//fmt.Printf("---------------%d-----------------\n", timeNow)
	pcb, value := findHighestPriorProcess(cpu)
	if !value {
		fmt.Printf("Cpu's processes_queue is null...")
		return
	}

	arrivalTime := pcb.arriveTime
	handleTime := timeNow

	changeCpuStatus(cpu, Run)
	fmt.Printf("handling process: %s at %d, which is arrived at %d, enter to the processes_queue...\n", pcb.name, handleTime, arrivalTime)

	fmt.Printf("handling process: %s...\n", pcb.name)
	stimulateCpuExecTime(pcb)
	if pcb.usedTime < pcb.execTime {
		fmt.Printf("process %s still need to exec %d time\n", pcb.name, pcb.execTime-pcb.usedTime)
		pcb.prior -= 1
		InQueue(cpu, pcb)
	} else {
		fmt.Printf("process:%s is finished\n", pcb.name)
		pcb.PCBStatus = Finish
		defer wg.Done()
	}
	changeCpuStatus(cpu, Wait)
	OutQueue(cpu, pcb)
	showCpuProcessQueue(cpu)
	fmt.Printf("\n")
}

func processIn(cpu *CpuProcessScheduler, pcb *PCB) {
	pcb.arriveTime = timeNow
	InQueue(cpu, pcb)
	wg.Add(1)
}

func main() {
	cpu := CpuProcessScheduler{
		processesQueue: nil,
		cpuStatus:      Wait,
	}

	go func() {
		for {
			value := <-ch
			if value == -1 {
				break
			}
			go CpuHandleProcess(&cpu)
		}
	}()

	pcb1 := PCB{name: "P1", prior: GenerateRandInt(20), arriveTime: 0, execTime: 10, usedTime: 0, PCBStatus: Wait}
	pcb2 := PCB{name: "P2", prior: GenerateRandInt(20), arriveTime: 0, execTime: 6, usedTime: 0, PCBStatus: Wait}
	pcb3 := PCB{name: "P3", prior: GenerateRandInt(20), arriveTime: 0, execTime: 4, usedTime: 0, PCBStatus: Wait}
	pcb4 := PCB{name: "P4", prior: GenerateRandInt(20), arriveTime: 0, execTime: 8, usedTime: 0, PCBStatus: Wait}
	pcb5 := PCB{name: "P5", prior: GenerateRandInt(20), arriveTime: 0, execTime: 3, usedTime: 0, PCBStatus: Wait}

	showPCB(pcb1)
	//showPCB(pcb2)
	//showPCB(pcb3)
	//showPCB(pcb4)
	//showPCB(pcb5)
	processIn(&cpu, &pcb1)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb2)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb3)
	time.Sleep(25 * time.Millisecond)
	processIn(&cpu, &pcb4)
	time.Sleep(25 * time.Millisecond)
	//fmt.Printf("\n\n\nwt? at time %d P2 is already In?---------------------------------------------------\n\n\n", timeNow)
	processIn(&cpu, &pcb5)

	wg.Wait()
}
