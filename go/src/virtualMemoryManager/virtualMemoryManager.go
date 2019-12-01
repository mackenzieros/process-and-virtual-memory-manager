package virtualMemoryManager

import (
	"fmt"
	"math"
	"os"
	"strconv"
	// DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
)

type VirtualMemoryManager struct {
	PhysicalMemory [524288]int
	PagingDisk     [1024][512]int
}

func printVirtualAddressComponents(virtualAddress string, segmentFrameNum int, pageTableAddr int, pageEntry int, pageFrameNum int, pageAddr int, physicalAddress int) {
	fmt.Printf("Virtual Address: %s\n", virtualAddress)
	fmt.Printf("Segment frame #: %d\n", segmentFrameNum)
	fmt.Printf("Page table address: %d\n", pageTableAddr)
	fmt.Printf("Page entry: %d\n", pageEntry)
	fmt.Printf("Page frame #: %d\n", pageFrameNum)
	fmt.Printf("Physical address: %d\n", physicalAddress)
	fmt.Println()
}

func TranslateVAs(vmm *VirtualMemoryManager, virtualAddresses []string) {
	aux_val := 511
	var s int
	var w int
	var p int
	var pw int
	var physicalAddress int
	for _, virtualAddress := range virtualAddresses {
		// vaComp := vaComponents{}
		va, err := strconv.Atoi(virtualAddress)

		if err != nil {
			fmt.Printf("Error converting string to int (virtual address): %s\n", err)
			return
		}

		s = va >> 18
		w = va & aux_val
		p = (va >> 9) & aux_val
		pw = va & 262143

		var result string
		if pw > vmm.PhysicalMemory[2*s] {
			fmt.Println("Error: virtual address outside segment boundary")
			result = "error"
		} else {
			absSegmentFrameNum := math.Abs(float64(vmm.PhysicalMemory[2*s+1]))
			segmentFrameNum := int(absSegmentFrameNum)
			pageTableAddr := segmentFrameNum * 512
			pageEntry := pageTableAddr + p
			pageFrameNum := vmm.PhysicalMemory[pageEntry]
			pageAddr := pageFrameNum * 512
			physicalAddress = pageAddr + w

			printVirtualAddressComponents(virtualAddress, segmentFrameNum, pageTableAddr, pageEntry, pageFrameNum, pageAddr, physicalAddress)
			result = strconv.Itoa(physicalAddress)
		}

		writeToOutput(result)
	}
}

func writeToOutput(content string) {
	f, err := os.OpenFile("../../../output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(content + " ")
	if err != nil {
		panic(err)
	}
}

func InitVirtualMemoryManager() *VirtualMemoryManager {
	var vmm VirtualMemoryManager
	return &vmm
}
