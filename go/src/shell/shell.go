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
		processManager.Reset(pm)
	case "rq":
		processManager.Request(pm, cmdNum1, cmdNum2)
	case "rl":
		processManager.Release(pm, cmdNum1, cmdNum2)
	default:
		fmt.Printf("Unable to process unknown command: %q.\n", cmdSlice[0])
	}
}

func RunShell() {
	var pm = &processManager.ProcessManager{}

	filename := os.Args[1]
	f, _ := os.Open(filename)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := scanner.Text()
		cmdSlice := strings.Fields(cmd)
		if len(cmdSlice) == 0 {
			continue
		}
		processCmds(cmdSlice, pm)
	}
}

func main() {
	RunShell()
}
