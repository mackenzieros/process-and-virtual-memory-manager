package virtualMemoryManager

import (
// DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
)

type VirtualMemoryManager struct {
	PhysicalMemory [524288]int
	PagingDisk     [1024][512]int
}

func InitVirtualMemoryManager() *VirtualMemoryManager {
	var vmm VirtualMemoryManager
	return &vmm
}
