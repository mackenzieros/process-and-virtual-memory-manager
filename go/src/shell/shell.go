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
	var cmdNum int
	var err error
	if len(cmdSlice) > 1 {
		cmdNum, err = strconv.Atoi(cmdSlice[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	switch cmdSlice[0] {
	case "cr":
		processManager.Create(pm)
		fmt.Println(pm)
	case "de":
		processManager.Destroy(pm, cmdNum)
	case "to":
		processManager.Timeout(pm)
	case "in":
		RunShell()
	default:
		fmt.Printf("Unable to process unknown command: %q.", cmdSlice[0])
	}
}

func RunShell() {
	var pm = processManager.InitProcessManager()
	processManager.Create(&pm)

	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Print("Enter command: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		cmdSlice := strings.Fields(cmd)
		if cmdSlice[0] == "q" {
			break
		}
		processCmds(cmdSlice, &pm)
	}
}

func main() {
	RunShell()
}
