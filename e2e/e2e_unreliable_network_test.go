package e2e

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestE2E_Unreliable_Network(t *testing.T) {
	nodesCount := 20 + rand.Intn(11) // vary nodes [20,30]
	maxFaulty := nodesCount/3 - 1
	maxHeight := uint64(200)
	currentHeight := uint64(0)
	jitterMax := 500 * time.Millisecond
	hook := newPartitionTransport(jitterMax)
	c := newPBFTCluster(t, "network_unreliable", "prt", nodesCount, hook)
	fmt.Printf("Starting cluster with %d nodes, max faulty %d.\n", nodesCount, maxFaulty)
	c.Start()

	for {
		currentHeight = currentHeight + 10
		var minorityPartition []string
		var majorityPartition []string
		// create 2 partition with random number of nodes
		// minority with no more that maxFaulty and majority with rest of the nodes
		pSize := 1 + rand.Intn(maxFaulty)
		for i := 0; i < pSize; i++ {
			minorityPartition = append(minorityPartition, "prt_"+strconv.Itoa(i))
		}
		for i := pSize; i < nodesCount; i++ {
			majorityPartition = append(majorityPartition, "prt_"+strconv.Itoa(i))
		}
		fmt.Printf("Partitions ratio %d/%d\n", len(majorityPartition), len(minorityPartition))

		hook.Partition(minorityPartition, majorityPartition)
		c.WaitForHeight(currentHeight, 10*time.Minute, false, majorityPartition)
		// randomly drop if possible nodes from the partition pick one number
		dropN := rand.Intn(maxFaulty - pSize + 1)
		fmt.Printf("Dropping: %v nodes.\n", dropN)

		currentHeight += 10
		// stop N nodes from majority partition
		for i := 0; i < dropN; i++ {
			c.nodes["prt_"+strconv.Itoa(pSize+i)].Stop()
		}

		var runningMajorityNodes []string
		var stoppedNodes []string
		for _, v := range c.nodes {
			if v.IsRunning() {
				for _, bp := range majorityPartition {
					if bp == v.name { // is part of the bigPartition
						runningMajorityNodes = append(runningMajorityNodes, v.name)
					}
				}
			} else {
				stoppedNodes = append(stoppedNodes, v.name)
			}
		}
		// check all running nodes in majority partition for the block height
		c.WaitForHeight(currentHeight, 10*time.Minute, false, runningMajorityNodes)
		// restart network for this iteration
		hook.Reset()
		for _, stopped := range stoppedNodes {
			c.nodes[stopped].Start()
		}

		if currentHeight >= maxHeight {
			break
		}
	}
	hook.Reset()
	// all nodes in the network should be synced after starting all nodes and partition restart
	c.WaitForHeight(maxHeight+20, 20*time.Minute, false)
	c.Stop()
}
