package processManager

import (
	"fmt"
	"os"

	DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
)

type pcb struct {
	state     int
	parent    int
	children  *DoublyLinkedList.List
	resources *DoublyLinkedList.List
	priority  int
	index     int
	blockedOn int
}

type rcb struct {
	state    int
	waitlist *DoublyLinkedList.List
}

func compareByPriority(a, b interface{}) int {
	c1 := a.(*pcb)
	c2 := b.(*pcb)

	switch {
	case c1.priority < c2.priority:
		return 1
	case c1.priority > c2.priority:
		return -1
	default:
		return 0
	}
}

func findAvailableProcess(pcbArr [16]*pcb) int {
	for i := 0; i < len(pcbArr); i++ {
		if pcbArr[i] == nil {
			return i
		}
	}
	return -1
}

func Create(pm *ProcessManager, priority int) {
	if pm.pcbList[15] != nil {
		fmt.Printf("Error: process list capacity at maximum\n")
		os.Exit(1)
	}

	if priority != 0 && priority != 1 && priority != 2 {
		fmt.Printf("Error: priority can only be values 0, 1, 2\n")
		os.Exit(1)
	}

	parentProcess, withinBounds := pm.readyList.Get(0)

	// Determine parent (-1 if root process)
	var parent = -1
	if parentProcess != nil && withinBounds {
		parent = parentProcess.(*pcb).index // type assertion
	}

	// Allocate new PCB with necessary default values
	freeIndex := findAvailableProcess(pm.pcbList)
	var newPcb = &pcb{1, parent, DoublyLinkedList.New(), DoublyLinkedList.New(), priority, freeIndex, -1}
	pm.pcbList[freeIndex] = newPcb

	fmt.Printf("Process %d created\n", freeIndex)

	// Add to ready list
	pm.readyList.Add(newPcb)

	// If root, no parent, so no need to add to a parent list
	if parent == -1 {
		return
	}

	pm.pcbList[parent].children.Add(newPcb)
	pm.pcbList[parent].children.Values()

	pm.readyList.Sort(compareByPriority)
	scheduler(pm)

	return
}

func canDelete(pm *ProcessManager, processInQuestion *pcb, childProcess *pcb) bool {
	parentIndex := childProcess.parent
	if parentIndex == -1 {
		return false
	}
	currProcess := pm.pcbList[parentIndex]
	for currProcess != nil {
		if processInQuestion == currProcess {
			return true
		}
		if currProcess.parent == -1 {
			return false
		}
		currProcess = pm.pcbList[currProcess.parent]
	}
	return false
}

func findIndexOfResourceToRelease(pm *ProcessManager, resourceInQuestion *rcb) int {
	for i := 0; i < len(pm.rcbList); i++ {
		if resourceInQuestion == pm.rcbList[i] {
			return i
		}
	}
	return -1
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

	processToDel := pm.pcbList[processIndex]

	// Check if current process is parent of process to destroy
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)
	if !canDelete(pm, currentProcess, processToDel) {
		fmt.Printf("Error: current process: %d is not a parent of process: %d\n", currentProcess.index, processToDel.index)
		os.Exit(1)
	}

	numProcessesDestroyed := 0
	// Destroy all children
	childrenList := processToDel.children
	for i := 0; i < childrenList.Size(); i++ {
		child, _ := childrenList.Get(i)
		numProcessesDestroyed += Destroy(pm, child.(*pcb).index)
	}

	// Remove from parent's children list
	parentIndex := processToDel.parent
	parentChildrenList := pm.pcbList[parentIndex].children
	indexOfProcessInChildrenList := parentChildrenList.IndexOf(processToDel)
	parentChildrenList.Remove(indexOfProcessInChildrenList)

	// Remove from Ready List or Waiting List
	var listToRemoveFrom *DoublyLinkedList.List
	if processToDel.state == 1 {
		listToRemoveFrom = pm.readyList
	} else if processToDel.state == 0 {
		listToRemoveFrom = pm.rcbList[processToDel.blockedOn].waitlist
	}
	indexOfProcessInList := listToRemoveFrom.IndexOf(processToDel)
	listToRemoveFrom.Remove(indexOfProcessInList)

	// Release all resources held by this process
	for i := 0; i < processToDel.resources.Size(); i++ {
		resource, _ := processToDel.resources.Get(i)
		indexOfResouceToRelease := findIndexOfResourceToRelease(pm, resource.(*rcb))
		if indexOfResouceToRelease == -1 {
			fmt.Println("Error: could not find resource to release in Destroy...")
			os.Exit(1)
		}
		resourceToRelease := pm.rcbList[indexOfResouceToRelease]
		unblockProcessOnRelease(pm, resourceToRelease)
	}

	// Free PCB from PCB list (removes from index)
	pm.pcbList[processIndex] = nil

	pm.readyList.Sort(compareByPriority)
	scheduler(pm)

	return numProcessesDestroyed
}

func Request(pm *ProcessManager, requestIndex int) {
	if requestIndex < 0 || requestIndex > len(pm.rcbList) {
		fmt.Printf("Error: request index: %d is out of range\n", requestIndex)
		os.Exit(1)
	}

	// Get requested resource and currently running process
	resourceToRequest := pm.rcbList[requestIndex]
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)

	// Root process is not allowed to request any resources
	if currentProcess.index == 0 {
		fmt.Println("Error: parent process not allowed to request any resources.")
		os.Exit(1)
	}

	if resourceToRequest == nil {
		// Resource not yet allocated, allocate it
		resourceToRequest = &rcb{0, DoublyLinkedList.New()}
		pm.rcbList[requestIndex] = resourceToRequest
	}

	currentlyHeldResource, _ := currentProcess.resources.Get(0)
	if currentlyHeldResource != nil && (currentlyHeldResource == resourceToRequest) {
		fmt.Printf("Error: cannot request resource %d. Already holding.\n", requestIndex)
		os.Exit(1)
	}

	if resourceToRequest.state == 0 {
		// Allocate free resource
		resourceToRequest.state = 1
		currentProcess.resources.Append(resourceToRequest)
		fmt.Printf("Resource %d allocated\n", requestIndex)
	} else {
		// Block current process
		currentProcess.state = 1
		currentProcess.blockedOn = requestIndex
		pm.readyList.Remove(0)
		resourceToRequest.waitlist.Append(currentProcess)
		fmt.Printf("Process %d blocked\n", currentProcess.index)
		scheduler(pm)
	}
}

func unblockProcessOnRelease(pm *ProcessManager, resourceToRelease *rcb) {
	// Un-block process on resource's waitlist and move it to the ready list
	unblockedProcessInterface, _ := resourceToRelease.waitlist.Get(0)
	resourceToRelease.waitlist.Remove(0)
	unblockedProcess := unblockedProcessInterface.(*pcb)

	pm.readyList.Append(unblockedProcess)

	unblockedProcess.blockedOn = -1
	unblockedProcess.state = 0
	unblockedProcess.resources.Append(resourceToRelease)
}

func Release(pm *ProcessManager, releaseIndex int) {
	if releaseIndex < 0 || releaseIndex > len(pm.rcbList) {
		fmt.Printf("Release index: %d is out of range\n", releaseIndex)
		return
	}

	// Get resource to release and currently running process
	resourceToRelease := pm.rcbList[releaseIndex]
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)

	// Check if current process is holding the resource to release
	currentlyHeldResource, _ := currentProcess.resources.Get(0)
	if currentlyHeldResource == nil || (currentlyHeldResource != resourceToRelease) {
		fmt.Printf("Error: cannot release resource %d. Not holding it.\n", releaseIndex)
		os.Exit(1)
	}
	// Make sure current process is not blocked on a release
	currentProcess.blockedOn = -1

	// Remove resource from currently running process' resource list
	indexOfResource := currentProcess.resources.IndexOf(resourceToRelease)
	currentProcess.resources.Remove(indexOfResource)

	if resourceToRelease.waitlist.Empty() {
		// No waiting processes, so set to free
		resourceToRelease.state = 0
	} else {
		unblockProcessOnRelease(pm, resourceToRelease)
	}
	fmt.Printf("Resource %d released\n", releaseIndex)
	pm.readyList.Sort(compareByPriority)
	scheduler(pm)
}

func scheduler(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	if currRunningProcess == nil {
		fmt.Println("Error: all processes blocked.")
		os.Exit(1)
	} else {
		fmt.Printf("Process %d running\n", currRunningProcess.(*pcb).index)
	}
}

func Timeout(pm *ProcessManager) {
	currRunningProcessInterface, _ := pm.readyList.Get(0)
	currRunningProcess := currRunningProcessInterface.(*pcb)
	pm.readyList.Remove(0)
	pm.readyList.Add(currRunningProcess)
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
