package shell

import (
	"fmt"
	"bufio"
	"os"
)

func RunShell() {
	fmt.Printf("hi")
	// var inputs []string

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter command: ")
	cmd, _ := reader.ReadString('\n')
	fmt.Printf(cmd)
}