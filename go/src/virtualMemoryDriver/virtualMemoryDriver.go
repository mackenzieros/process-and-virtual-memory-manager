package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	virtualMemoryManager "virtualMemoryManager"
)

func processSegmentTableLine(vmm *virtualMemoryManager.VirtualMemoryManager, stLineSliced []string) {
	// Iterate through the line and init pm
	var segmentTriple []int
	for index, value := range stLineSliced {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Error converting string to int (segment table processing): %s\n", err)
			return
		}

		segmentTriple = append(segmentTriple, intVal)

		// We have not yet reached three elements in the segment triple
		if (index+1)%3 != 0 || len(segmentTriple) != 3 {
			continue
		}

		// Parse triple and place values into PM
		segmentNum := segmentTriple[0]
		segmentLen := segmentTriple[1]
		frameNum := segmentTriple[2]

		(*vmm).PhysicalMemory[2*segmentNum] = segmentLen
		(*vmm).PhysicalMemory[2*segmentNum+1] = frameNum
		segmentTriple = segmentTriple[:0]
	}
}

func processPageTableLine(vmm *virtualMemoryManager.VirtualMemoryManager, ptLineSliced []string) {
	// Iterate through the line and init pm
	var pageTriple []int
	for index, value := range ptLineSliced {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Error converting string to int (page table processing): %s\n", err)
			return
		}

		pageTriple = append(pageTriple, intVal)

		// We have not yet reached three elements in the page triple
		if (index+1)%3 != 0 || len(pageTriple) != 3 {
			continue
		}

		// Parse triple and place values into PM
		segmentNum := pageTriple[0]
		pageNum := pageTriple[1]
		frameNum := pageTriple[2]

		absFrameNum := math.Abs(float64((*vmm).PhysicalMemory[2*segmentNum+1]))
		pageFrameNum := int(absFrameNum)*512 + pageNum
		(*vmm).PhysicalMemory[pageFrameNum] = frameNum
		pageTriple = pageTriple[:0]
	}
}

func initializePhysicalMemory(vmm *virtualMemoryManager.VirtualMemoryManager, initFilepath string) {
	f, _ := os.Open(initFilepath)
	scanner := bufio.NewScanner(f)

	// Read the segment table init line (s z f, where s=segment#, z=segment length, f=frame#)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Err reading segment table init line: %s\n", err)
		return
	}

	// Slicify segment table init line
	stInp := scanner.Text()
	stLine := strings.Fields(stInp)

	processSegmentTableLine(vmm, stLine)

	// Read the page table init line (s p f, where s=segment#, p=page# length, f=frame#)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Err reading segment table init line: %s\n", err)
		return
	}

	// Slicify segment table init line
	ptInp := scanner.Text()
	ptLine := strings.Fields(ptInp)

	processPageTableLine(vmm, ptLine)
}

func readVAs(vaFilepath string) []string {
	f, _ := os.Open(vaFilepath)
	scanner := bufio.NewScanner(f)

	// Read the virtual addresses
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Err reading virtual addresses: %s\n", err)
	}

	// Slicify virtual addresses
	inp := scanner.Text()
	virtualAddresses := strings.Fields(inp)

	return virtualAddresses
}

func RunDriver() {
	pmInitFilename := os.Args[1]
	var vmm = virtualMemoryManager.InitVirtualMemoryManager()
	initializePhysicalMemory(vmm, pmInitFilename)
	// fmt.Println(vmm.PhysicalMemory[vmm.PhysicalMemory[2*8+1]*512+0])
	vaToTranslateFilename := os.Args[2]
	virtualAddresses := readVAs(vaToTranslateFilename)
	virtualMemoryManager.TranslateVAs(vmm, virtualAddresses)
}

func main() {
	RunDriver()
}
