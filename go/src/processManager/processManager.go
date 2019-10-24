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

func findAvailableProcess(pcbArr [16]*pcb) int {
	for i := 0; i < len(pcbArr); i++ {
		if pcbArr[i] == nil {
			return i
		}
	}
	return -1
}

func Create(pm *ProcessManager) {
	parentProcess, withinBounds := pm.readyList.Get(0)

	// Determine parent (-1 if root process)
	var parent = -1
	if parentProcess != nil && withinBounds {
		parent = parentProcess.(*pcb).index // type assertion
	}

	// Allocate new PCB with necessary default values
	// var latestIndex = len(pm.pcbList)
	freeIndex := findAvailableProcess(pm.pcbList)
	var newPcb = pcb{1, parent, DoublyLinkedList.New(), DoublyLinkedList.New(), freeIndex, -1}
	pm.pcbList[freeIndex] = &newPcb

	fmt.Printf("Process %d created\n", freeIndex)

	// Add to ready list
	pm.readyList.Add(&newPcb)

	// If root, no parent, so no need to add to a parent list
	if parent == -1 {
		return
	}

	pm.pcbList[parent].children.Add(&newPcb)
	pm.pcbList[parent].children.Values()

	return
}

func Destroy(pm *ProcessManager, processIndex int) int {
	if processIndex < 0 || processIndex > len(pm.pcbList) {
		fmt.Printf("Process index to destroy: %d is out of range\n", processIndex)
		return -1
	}

	if pm.pcbList[processIndex] == nil {
		fmt.Printf("Process at index %d is nil\n", processIndex)
		return -1
	}

	process := pm.pcbList[processIndex]

	numProcessesDestroyed := 0
	// Destroy all children
	childrenList := process.children
	for i := 0; i < childrenList.Size(); i++ {
		child, _ := childrenList.Get(i)
		numProcessesDestroyed += Destroy(pm, child.(*pcb).index)
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
	pm.pcbList[processIndex] = nil

	return numProcessesDestroyed
}

func Request(pm *ProcessManager, requestIndex int) {
	if requestIndex < 0 || requestIndex > len(pm.rcbList) {
		fmt.Printf("Request index: %d is out of range\n", requestIndex)
		return
	}

	resourceToRequest := pm.rcbList[requestIndex]
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)

	if resourceToRequest == nil {
		resourceToRequest = &rcb{0, DoublyLinkedList.New()}
		pm.rcbList[requestIndex] = resourceToRequest
	}

	if resourceToRequest.state == 0 {
		resourceToRequest.state = 1
		currentProcess.resources.Append(&resourceToRequest)
		fmt.Printf("Resource %d allocated\n", requestIndex)
	} else {
		currentProcess.state = 1
		currentProcess.blockedOn = requestIndex
		pm.readyList.Remove(0)
		resourceToRequest.waitlist.Append(&currentProcess)
		fmt.Printf("Process %d blocked\n", currentProcess.index)
		scheduler(pm)
	}
}

func scheduler(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	if currRunningProcess == nil {
		fmt.Println("All processes blocked.")
	} else {
		fmt.Printf("Process %d running\n", currRunningProcess.(*pcb).index)
	}
}

func Timeout(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	pm.readyList.Remove(0)
	pm.readyList.Add(&currRunningProcess)
	scheduler(pm)
}

type ProcessManager struct {
	pcbList   [16]*pcb
	rcbList   [4]*rcb
	readyList *DoublyLinkedList.List
}

func InitProcessManager() ProcessManager {
	fmt.Printf("Initializing process manager...\n")
	var processManager ProcessManager
	processManager.readyList = DoublyLinkedList.New()
	fmt.Printf("Process manager initialized!\n")
	return processManager
}
