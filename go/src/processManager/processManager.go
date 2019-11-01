package processManager

import (
	"fmt"

	DoublyLinkedList "github.com/emirpasic/gods/lists/doublylinkedlist"
)

type resourcesHolding struct {
	resource *rcb
	numUnits int
}

type pcb struct {
	state     int
	parent    int
	children  *DoublyLinkedList.List
	resources *DoublyLinkedList.List
	priority  int
	index     int
	blockedOn int
}

type resourcesNeeded struct {
	process  *pcb
	numUnits int
}

type rcb struct {
	state     int
	waitlist  *DoublyLinkedList.List
	inventory int
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
		fmt.Printf("-1 ")
		return
	}

	if priority != 0 && priority != 1 && priority != 2 {
		fmt.Printf("-1 ")
		return
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

	// Add to ready list
	pm.readyList.Add(newPcb)

	// If root, no parent, so no need to add to a parent list
	if parent == -1 {
		scheduler(pm)
		return
	}

	pm.pcbList[parent].children.Add(newPcb)

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

func findProcessInWaitlist(waitlist *DoublyLinkedList.List, process *pcb) int {
	for i := 0; i < waitlist.Size(); i++ {
		needyProcessObj, _ := waitlist.Get(i)
		if needyProcessObj.(*resourcesNeeded).process == process {
			return i
		}
	}
	return -1
}

func destroyAuxillary(
	pm *ProcessManager,
	processIndex int,
	childrenList *DoublyLinkedList.List,
	processToDel *pcb,
	numProcessesDestroyed int,
) int {
	for {
		child, _ := childrenList.Get(0)
		if child == nil {
			break
		}
		numProcessesDestroyed += destroyAuxillary(
			pm,
			child.(*pcb).index,
			child.(*pcb).children,
			child.(*pcb),
			numProcessesDestroyed,
		)
	}

	// Remove from parent's children list
	parentIndex := processToDel.parent
	parentChildrenList := pm.pcbList[parentIndex].children
	indexOfProcessInChildrenList := parentChildrenList.IndexOf(processToDel)
	parentChildrenList.Remove(indexOfProcessInChildrenList)

	// Release all resources held by this process
	for i := 0; i < processToDel.resources.Size(); i++ {
		resourceObj, _ := processToDel.resources.Get(i)
		indexOfResourceToRelease := findIndexOfResourceToRelease(pm, resourceObj.(*resourcesHolding).resource)
		if indexOfResourceToRelease == -1 {
			fmt.Printf("-1 ")
			return 0
		}
		resourceToRelease := pm.rcbList[indexOfResourceToRelease]
		updateProcessesOnRelease(pm, resourceToRelease, resourceObj.(*resourcesHolding).numUnits)
	}

	// Remove from Ready List or Waiting List
	var listToRemoveFrom *DoublyLinkedList.List
	var indexOfProcessInList int
	if processToDel.state == 1 {
		listToRemoveFrom = pm.readyList
		indexOfProcessInList = listToRemoveFrom.IndexOf(processToDel)
	} else if processToDel.state == 0 {
		listToRemoveFrom = pm.rcbList[processToDel.blockedOn].waitlist
		indexOfProcessInList = findProcessInWaitlist(listToRemoveFrom, processToDel)
	}
	listToRemoveFrom.Remove(indexOfProcessInList)

	// Free PCB from PCB list (removes from index)
	pm.pcbList[processIndex] = nil

	return numProcessesDestroyed
}

func Destroy(pm *ProcessManager, processIndex int) int {
	if processIndex < 0 || processIndex > len(pm.pcbList) {
		fmt.Printf("-1 ")
		return 0
	}

	if pm.pcbList[processIndex] == nil {
		fmt.Printf("-1 ")
		return 0
	}

	processToDel := pm.pcbList[processIndex]
	// Check if current process is parent of process to destroy
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)
	if !canDelete(pm, currentProcess, processToDel) {
		fmt.Printf("-1 ")
		return 0
	}

	// Destroy all children
	childrenList := processToDel.children
	numProcessesDestroyed := destroyAuxillary(
		pm,
		processIndex,
		childrenList,
		processToDel,
		0,
	)

	pm.readyList.Sort(compareByPriority)
	scheduler(pm)

	return numProcessesDestroyed
}

func findIndexOfHeldResource(resources *DoublyLinkedList.List, resourceInQuestion *rcb) int {
	for i := 0; i < resources.Size(); i++ {
		heldResourceObj, _ := resources.Get(i)
		if resourceInQuestion == heldResourceObj.(*resourcesHolding).resource {
			return i
		}
	}
	return -1
}

func Request(pm *ProcessManager, requestIndex int, numUnits int) {
	if requestIndex < 0 || requestIndex > len(pm.rcbList) {
		fmt.Printf("-1 ")
		return
	}

	// Get requested resource and currently running process
	resourceToRequest := pm.rcbList[requestIndex]
	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)

	// Root process is not allowed to request any resources
	if currentProcess.index == 0 {
		fmt.Printf("-1 ")
		return
	}

	currentlyHeldResourceObj, _ := currentProcess.resources.Get(0)
	// Check if number to release does not exceed amount currently held
	amtCurrentlyHeld := resourceToRequest.inventory - resourceToRequest.state
	// Check so that current resource cannot request more resources than allowed
	if currentlyHeldResourceObj != nil && (currentlyHeldResourceObj.(*resourcesHolding).resource == resourceToRequest) &&
		numUnits > amtCurrentlyHeld {
		fmt.Printf("-1 ")
		return
	}

	if resourceToRequest.state-numUnits >= 0 {
		// Allocate free resource
		resourceIndex := findIndexOfHeldResource(currentProcess.resources, resourceToRequest)
		if resourceIndex == -1 {
			currentProcess.resources.Append(&resourcesHolding{resourceToRequest, numUnits})
		} else {
			alreadyOwnedResource, _ := currentProcess.resources.Get(resourceIndex)
			alreadyOwnedResource.(*resourcesHolding).numUnits += numUnits
		}
		resourceToRequest.state -= numUnits
	} else {
		// Block current process
		currentProcess.state = 0
		currentProcess.blockedOn = requestIndex
		pm.readyList.Remove(0)
		resourceToRequest.waitlist.Append(&resourcesNeeded{currentProcess, numUnits})
	}
	pm.readyList.Sort(compareByPriority)
	scheduler(pm)
}

func updateProcessesOnRelease(pm *ProcessManager, resourceToRelease *rcb, numUnits int) {
	resourceToRelease.state += numUnits
	// Un-block process on resource's waitlist and move it to the ready list
	for i := 0; i < resourceToRelease.waitlist.Size(); i++ {
		processToUnblockInterface, _ := resourceToRelease.waitlist.Get(i)
		processToUnblockInfo := processToUnblockInterface.(*resourcesNeeded)

		if resourceToRelease.state >= processToUnblockInfo.numUnits {
			resourceToRelease.waitlist.Remove(i)
			processToUnblockInfo.process.blockedOn = -1
			processToUnblockInfo.process.state = 1
			pm.readyList.Add(processToUnblockInfo.process)
		}
	}
}

func Release(pm *ProcessManager, releaseIndex int, numUnits int) {
	if releaseIndex < 0 || releaseIndex > len(pm.rcbList) {
		fmt.Printf("-1 ")
		return
	}

	// Get resource to release and currently running process
	resourceToRelease := pm.rcbList[releaseIndex]

	currentProcessInterface, _ := pm.readyList.Get(0)
	currentProcess := currentProcessInterface.(*pcb)

	// Check if current process is holding the resource to release
	currentlyHeldResourceObj, _ := currentProcess.resources.Get(0)
	if currentlyHeldResourceObj == nil || (currentlyHeldResourceObj.(*resourcesHolding).resource != resourceToRelease) {
		fmt.Printf("-1 ")
		return
	}

	// Check if number to release does not exceed amount currently held
	if numUnits > currentlyHeldResourceObj.(*resourcesHolding).numUnits {
		fmt.Printf("-1 ")
		return
	}

	for i := 0; i < numUnits; i++ {
		if resourceToRelease.waitlist.Empty() {
			// No waiting processes, so set to free
			resourceToRelease.state++
		} else {
			updateProcessesOnRelease(pm, resourceToRelease, 1)
		}
		pm.readyList.Sort(compareByPriority)
	}

	// Remove resource from currently running process' resource list when resources run out
	indexOfResource := currentProcess.resources.IndexOf(currentlyHeldResourceObj.(*resourcesHolding))
	if currentlyHeldResourceObj.(*resourcesHolding).numUnits-numUnits == 0 {
		currentProcess.resources.Remove(indexOfResource)
	} else {
		currentlyHeldResourceObj.(*resourcesHolding).numUnits -= numUnits
	}
	scheduler(pm)
}

func scheduler(pm *ProcessManager) {
	currRunningProcess, _ := pm.readyList.Get(0)
	if currRunningProcess == nil {
		fmt.Printf("-1 ")
	} else {
		fmt.Printf("%d ", currRunningProcess.(*pcb).index)
	}
}

func Timeout(pm *ProcessManager) {
	currRunningProcessInterface, _ := pm.readyList.Get(0)
	currRunningProcess := currRunningProcessInterface.(*pcb)
	pm.readyList.Remove(0)
	pm.readyList.Add(currRunningProcess)
	pm.readyList.Sort(compareByPriority)
	scheduler(pm)
}

func Reset(pm *ProcessManager) {
	fmt.Println()
	InitProcessManager(pm)
}

type ProcessManager struct {
	pcbList   [16]*pcb
	rcbList   [4]*rcb
	readyList *DoublyLinkedList.List
}

func InitProcessManager(pm *ProcessManager) {
	var processManager ProcessManager
	processManager.rcbList[0] = &rcb{1, DoublyLinkedList.New(), 1}
	processManager.rcbList[1] = &rcb{1, DoublyLinkedList.New(), 1}
	processManager.rcbList[2] = &rcb{2, DoublyLinkedList.New(), 2}
	processManager.rcbList[3] = &rcb{3, DoublyLinkedList.New(), 3}
	processManager.readyList = DoublyLinkedList.New()
	Create(&processManager, 0)
	*pm = processManager
}
