package main

import (
	"bufio"
	"fmt"
	DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
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

		// If frame is allocated, set as free,
		// otherwise, save info on block data
		if frameNum*(-1) < 1 {
			frame, _ := vmm.FreeFrames.Get(frameNum)
			frame.(*virtualMemoryManager.DiskFrame).Free = 1
		} else {
			frame, _ := vmm.FreeFrames.Get(frameNum * (-1))
			frame.(*virtualMemoryManager.DiskFrame).Free = 0
			frame.(*virtualMemoryManager.DiskFrame).Index = frameNum * (-1)
			frame.(*virtualMemoryManager.DiskFrame).Block[0] = segmentLen
			frame.(*virtualMemoryManager.DiskFrame).Block[1] = frameNum * (-1)
		}
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
		pageFrameNum := pageTriple[2]

		rawFrameNum := (*vmm).PhysicalMemory[2*segmentNum+1]
		absFrameNum := math.Abs(float64(rawFrameNum))
		pageFrameNumIndex := int(absFrameNum)*512 + pageNum
		(*vmm).PhysicalMemory[pageFrameNumIndex] = pageFrameNum
		pageTriple = pageTriple[:0]

		// If frame is allocated, set as free,
		// otherwise, save info on block data
		if rawFrameNum*(-1) < 1 {
			frame, _ := vmm.FreeFrames.Get(rawFrameNum)
			frame.(*virtualMemoryManager.DiskFrame).Free = 1
		} else {
			frame, _ := vmm.FreeFrames.Get(rawFrameNum * (-1))
			frame.(*virtualMemoryManager.DiskFrame).Free = 0
			frame.(*virtualMemoryManager.DiskFrame).Index = rawFrameNum * (-1)
			frame.(*virtualMemoryManager.DiskFrame).Block[pageNum] = pageFrameNum
		}
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

	// Initialize all frames in Page Disk to be free except frames 0 and 1
	vmm.FreeFrames = DoublyLinkedList.New()
	for i := 0; i < 1024; i++ {
		freeVal := 0
		if i == 0 || i == 1 {
			freeVal = 1
		}

		var block [512]int
		vmm.FreeFrames.Add(&virtualMemoryManager.DiskFrame{Index: i, Free: freeVal, Block: block})
	}

	initializePhysicalMemory(vmm, pmInitFilename)
	vaToTranslateFilename := os.Args[2]
	virtualAddresses := readVAs(vaToTranslateFilename)
	virtualMemoryManager.TranslateVAs(vmm, virtualAddresses)
}

func main() {
	RunDriver()
}
