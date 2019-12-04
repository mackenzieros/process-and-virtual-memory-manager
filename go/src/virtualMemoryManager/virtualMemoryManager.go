package virtualMemoryManager

import (
	"fmt"
	DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
	"math"
	"os"
	"strconv"
)

type DiskFrame struct {
	Index int
	Free  int
	Block [512]int
}

type VirtualMemoryManager struct {
	PhysicalMemory [524288]int
	PagingDisk     [1024][512]int
	FreeFrames     *DoublyLinkedList.List
}

func printVirtualAddressComponents(virtualAddress string, segmentFrameNum int, pageTableAddr int, pageEntry int, pageFrameNum int) {
	fmt.Printf("Virtual Address: %s\n", virtualAddress)
	fmt.Printf("Segment frame #: %d\n", segmentFrameNum)
	fmt.Printf("Page table address: %d\n", pageTableAddr)
	fmt.Printf("Page entry: %d\n", pageEntry)
	fmt.Printf("Page frame #: %d\n", pageFrameNum)
	fmt.Println()
}

func allocateFreeFrame(vmm *VirtualMemoryManager) *DiskFrame {
	for i := 0; i < vmm.FreeFrames.Size(); i++ {
		frame, _ := vmm.FreeFrames.Get(i)
		if frame.(*DiskFrame).Free != 0 {
			continue
		}

		frame.(*DiskFrame).Free = 1
		return frame.(*DiskFrame)
	}
	return nil
}

func readBlock(vmm *VirtualMemoryManager, b int, f int) {
	pmIndex := f * 512
	index := int(math.Abs(float64(b)))
	frameIndex := int(math.Abs(float64(vmm.PhysicalMemory[index])))
	frameToRead, _ := vmm.FreeFrames.Get(frameIndex)
	// fmt.Println("frame to read: ", frameIndex)
	// fmt.Println("frame val: ", frameToRead.(*DiskFrame).Block[1])
	for i := 0; i < 511; i, pmIndex = i+1, pmIndex+1 {
		vmm.PhysicalMemory[pmIndex] = frameToRead.(*DiskFrame).Block[i]
	}
}

func TranslateVAs(vmm *VirtualMemoryManager, virtualAddresses []string) {
	aux_val := 511
	var s int
	var w int
	var p int
	var pw int
	var physicalAddress int
	for _, virtualAddress := range virtualAddresses {
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
		if pw >= vmm.PhysicalMemory[2*s] {
			fmt.Println("Error: virtual address outside segment boundary")
			result = "error"
			writeToOutput(result)
			continue
		}

		mult := 1
		if vmm.PhysicalMemory[2*s+1] < 0 {
			mult *= -1
		}

		if vmm.PhysicalMemory[2*s+1] < 0 {
			// page fault: Page table is not resident
			freeFrame := allocateFreeFrame(vmm)
			if freeFrame == nil {
				fmt.Println("No free frames found")
			}
			// fmt.Println("page table fault")
			readBlock(vmm, 2*s+1, freeFrame.Index)

			vmm.PhysicalMemory[2*s+1] = freeFrame.Index
		} else if vmm.PhysicalMemory[vmm.PhysicalMemory[2*s+1]*512+p] < 0 {
			// page fault: Page is not resident
			freeFrame := allocateFreeFrame(vmm)
			if freeFrame == nil {
				fmt.Println("No free frames found")
			}
			// fmt.Println("page fault")
			readBlock(vmm, 2*s+1, freeFrame.Index)

			vmm.PhysicalMemory[vmm.PhysicalMemory[2*s+1]*512+p] = freeFrame.Index
		}

		physicalAddress = vmm.PhysicalMemory[vmm.PhysicalMemory[2*s+1]*512+p]*512 + w
		fmt.Println("Physical Address: ", physicalAddress)
		result = strconv.Itoa(physicalAddress)

		writeToOutput(result)
	}
}

func writeToOutput(content string) {
	f, err := os.OpenFile("./output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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
