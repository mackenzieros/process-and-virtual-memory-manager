package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	processManager "processManager"
)

func processCmds(cmdSlice []string, pm *processManager.ProcessManager) {
	var cmdNum1 int
	var cmdNum2 int
	var err error
	if len(cmdSlice) > 1 {
		cmdNum1, err = strconv.Atoi(cmdSlice[1])
		if len(cmdSlice) > 2 {
			cmdNum2, err = strconv.Atoi(cmdSlice[2])
		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	switch cmdSlice[0] {
	case "cr":
		processManager.Create(pm, cmdNum1)
	case "de":
		processManager.Destroy(pm, cmdNum1)
	case "to":
		processManager.Timeout(pm)
	case "in":
		RunShell()
	case "rq":
		processManager.Request(pm, cmdNum1, cmdNum2)
	case "rl":
		processManager.Release(pm, cmdNum1, cmdNum2)
	default:
		fmt.Printf("Unable to process unknown command: %q.\n", cmdSlice[0])
	}
}

func RunShell() {
	var pm = processManager.InitProcessManager()
	processManager.Create(&pm, 0)
	fmt.Println()

	filename := os.Args[1]
	f, _ := os.Open(filename)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := scanner.Text()
		if len(cmd) == 0 {
			break
		}
		cmdSlice := strings.Fields(cmd)
		processCmds(cmdSlice, &pm)
		fmt.Println()
	}
	// for true {
	// 	fmt.Print("Enter command: ")
	// 	cmd, _ := reader.ReadString('\n')
	// 	cmd = strings.TrimSpace(cmd)
	// 	cmdSlice := strings.Fields(cmd)
	// 	if cmdSlice[0] == "q" {
	// 		break
	// 	}
	// 	processCmds(cmdSlice, &pm)
	// }
}

func main() {
	RunShell()
}
