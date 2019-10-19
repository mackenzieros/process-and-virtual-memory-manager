package processManager

import (
	"fmt"

	DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
)

type pcb struct {
	state     int
	parent    int
	children  *DoublyLinkedList.List
	resources *DoublyLinkedList.List
	index     int
	blockedOn int
}

type rcb struct {
	state    int
	waitlist *DoublyLinkedList.List
}

func Create(pm *ProcessManager) {
	parentProcess, withinBounds := pm.readyList.Get(0)

	// Determine parent (-1 if root process)
	var parent = -1
	if parentProcess != nil && withinBounds {
		parent = parentProcess.(pcb).index // type assertion
	}

	// Allocate new PCB with necessary default values
	var latestIndex = len(pm.pcbList)
	var newPcb = pcb{1, parent, DoublyLinkedList.New(), DoublyLinkedList.New(), latestIndex, -1}
	pm.pcbList = append(pm.pcbList, newPcb)

	fmt.Printf("Process %d created\n", latestIndex)

	// Add to ready list
	pm.readyList.Add(newPcb)

	// If root, no parent, so no need to add to a parent list
	if parent == -1 {
		return
	}

	pm.pcbList[parent].children.Add(newPcb)
	pm.pcbList[parent].children.Values()

	return
}

func Destroy(pm *ProcessManager, processIndex int) int {
	if processIndex < 0 || processIndex > len(pm.pcbList) {
		fmt.Printf("Process index to destroy: %d is out of range", processIndex)
	}

	process := pm.pcbList[processIndex]

	numProcessesDestroyed := 0
	// Destroy all children
	childrenList := process.children
	for i := 0; i < childrenList.Size(); i++ {
		child, _ := childrenList.Get(i)
		numProcessesDestroyed += Destroy(pm, child.(pcb).index)
	}

	// Remove from parent's children list
	parentIndex := process.parent
	parentChildrenList := pm.pcbList[parentIndex].children
	indexOfProcessInChildrenList := parentChildrenList.IndexOf(process)
	parentChildrenList.Remove(indexOfProcessInChildrenList)

	// Remove from Ready List or Waiting List
	var listToRemoveFrom *DoublyLinkedList.List
	if process.state == 1 {
		listToRemoveFrom = pm.readyList
	} else if process.state == 0 {
		listToRemoveFrom = pm.rcbList[process.blockedOn].waitlist
	}
	indexOfProcessInList := listToRemoveFrom.IndexOf(process)
	listToRemoveFrom.Remove(indexOfProcessInList)

	// Free PCB from PCB list (removes from index)
	pm.pcbList = append(pm.pcbList[:processIndex], pm.pcbList[processIndex+1:]...)

	return numProcessesDestroyed
}

func scheduler(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	fmt.Printf("Process %d running\n", currRunningProcess.(pcb).index)
}

func Timeout(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	pm.readyList.Remove(0)
	pm.readyList.Add(currRunningProcess)
	scheduler(pm)
}

type ProcessManager struct {
	pcbList   []pcb
	rcbList   []rcb
	readyList *DoublyLinkedList.List
}

func InitProcessManager() ProcessManager {
	fmt.Printf("Initializing process manager...\n")
	var processManager ProcessManager
	processManager.readyList = DoublyLinkedList.New()
	fmt.Printf("Process manager initialized!\n")
	return processManager
}
