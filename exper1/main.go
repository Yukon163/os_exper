package main

import (
	"fmt"
	"math/rand"
	"time"
)

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
	return rand.Intn(100)
}

type PCB struct {
	name        string
	prior       int // 优先级 优先数？
	arrive_time int64
	exec_time   int64
	used_time   int64
	PCBStatus   status
}

func showPCB(pcb PCB) {
	fmt.Printf("Process Name:%s, Status: %s\n", pcb.name, pcb.PCBStatus)
}

type CPU_Process_Scheduler struct {
	processes_queue []PCB
	cpuStatus       status
}

func InQueue(cpu *CPU_Process_Scheduler, pcb PCB) {
	cpu.processes_queue = append([]PCB{pcb}, cpu.processes_queue...)
}

func OutQueue(cpu *CPU_Process_Scheduler, pcb PCB) {
	for i, process := range cpu.processes_queue {
		if process.name == pcb.name {
			cpu.processes_queue = append(cpu.processes_queue[:i], cpu.processes_queue[i+1:]...)
			fmt.Printf("Process %s removed from queue\n", pcb.name)
			break
		}
	}
}

func getCpuProcessQueue(cpu *CPU_Process_Scheduler) {
	for i, pcb := range cpu.processes_queue {
		fmt.Printf("Process %d: %s\n", i, pcb.name)
	}
}

func findHighestPriorProcess(cpu *CPU_Process_Scheduler) PCB {
	if len(cpu.processes_queue) == 0 {
		fmt.Printf("No processes queue\n")
	}

	highestPriorProcess := cpu.processes_queue[0]
	for _, process := range cpu.processes_queue {
		if process.prior > highestPriorProcess.prior {
			highestPriorProcess = process
		}
	}
	return highestPriorProcess
}

func showCPU(cpu *CPU_Process_Scheduler) {
	fmt.Printf("CPUStatus: %s\n", cpu.cpuStatus)
}

func CPU_handle_process(cpu *CPU_Process_Scheduler, pcb PCB) {
	time.Sleep(1) // 模拟cpu执行进程花费的时间片
	pcb.used_time += 1
	if pcb.used_time < pcb.exec_time {
		fmt.Printf("process still need to exec %d time\n", (pcb.exec_time - pcb.used_time))
		pcb.prior -= 1
		OutQueue(cpu, pcb)
		process_in(cpu, pcb)
	} else {
		fmt.Printf("process:%s is finished\n", pcb.name)
		pcb.PCBStatus = Finish
	}
}

func process_in(cpu *CPU_Process_Scheduler, pcb PCB) {
	pcb.arrive_time = time.Now().UnixNano()
	if cpu.cpuStatus == Wait && pcb.PCBStatus == Wait {
		arrivalTime := time.Unix(0, pcb.arrive_time).Format("2006-01-02 15:04:05")
		fmt.Printf("CPU Status: %s->Run\nprocess: %s arrive at %s, enter to the processes_queue...\n", cpu.cpuStatus, pcb.name, arrivalTime)
		cpu.cpuStatus = Run
		cpu.processes_queue = append([]PCB{pcb}, cpu.processes_queue...) // 用于在开头插入pcb以模拟队列
		processToHandle := findHighestPriorProcess(cpu)
		fmt.Printf("start process: %s...\n", processToHandle.name)
		go CPU_handle_process(cpu, processToHandle)
	} else if cpu.cpuStatus == Wait && cpu.processes_queue != nil {
		processToHandle := findHighestPriorProcess(cpu)
		fmt.Printf("start process: %s...\n", processToHandle.name)
		CPU_handle_process(cpu, processToHandle)
	} else if cpu.cpuStatus == Run {
		fmt.Printf("CPU Status: %s...\nprocess: %s enter to the processes_queue...\n", cpu.cpuStatus, pcb.name)
		cpu.processes_queue = append(cpu.processes_queue, pcb)
		fmt.Printf("Now, the CPU's process_queue is:\n")
		getCpuProcessQueue(cpu)
	} else {
		fmt.Printf("wt?cpu seems to be wrong, just check it's status\n")
	}
}

func scheduleProcesses(cpu *CPU_Process_Scheduler, pcbs []PCB) {
	for _, pcb := range pcbs {
		process_in(cpu, pcb)
	}
}

func main() {
	cpu := CPU_Process_Scheduler{
		processes_queue: nil,
		cpuStatus:       Wait,
	}
	pcb1 := PCB{
		name:        "process1",
		prior:       GenerateRandInt(100),
		arrive_time: 0,
		exec_time:   10,
		used_time:   0,
		PCBStatus:   Wait,
	}
	showCPU(&cpu)
	showPCB(pcb1)
	process_in(&cpu, pcb1)

}
