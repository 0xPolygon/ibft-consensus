package e2e

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestE2E_NodeDrop(t *testing.T) {
	c := newPBFTCluster(t, "node_drop", "ptr", 5)
	c.Start()

	// wait for two heights and stop node 1
	c.WaitForHeight(2, 1*time.Minute, false)

	c.StopNode("ptr_0")
	c.WaitForHeight(15, 1*time.Minute, false, generateNodeNames(1, 4, "ptr_"))
}

func TestE2E_NetworkChurn(t *testing.T) {
	rand.Seed(time.Now().Unix())
	nodeCount := 20
	const prefix = "ptr_"
	c := newPBFTCluster(t, "network_churn", "ptr", nodeCount)
	c.Start()
	runningNodeCount := nodeCount
	// randomly stop nodes every 3 seconds
	executeInTimerAndWait(3*time.Second, 30*time.Second, func(_ time.Duration) {
		nodeNo := rand.Intn(nodeCount)
		nodeID := prefix + strconv.Itoa(nodeNo)
		node := c.nodes[nodeID]
		if node.IsRunning() && runningNodeCount > int(math.Ceil(float64(nodeCount)/3.0))-1 {
			// node is running
			c.StopNode(nodeID)
			runningNodeCount--
		} else if !node.IsRunning() {
			// node is not running
			c.StartNode(nodeID)
			runningNodeCount++
		}
	})

	// get all running nodes after random drops
	var runningNodes []string
	var stoppedNodes []string
	for _, v := range c.nodes {
		if v.IsRunning() {
			runningNodes = append(runningNodes, v.name)
		} else {
			stoppedNodes = append(stoppedNodes, v.name)
		}
	}

	// all stopped nodes are stuck
	c.IsStuck(30*time.Second, stoppedNodes)

	// all running nodes must have the same height
	c.WaitForHeight(15, 1*time.Minute, false, runningNodes)

	// start rest of the nodes
	for _, v := range c.nodes {
		if !v.IsRunning() {
			v.Start()
			runningNodes = append(runningNodes, v.name)
		}
	}

	// all nodes must sync and have same height
	c.WaitForHeight(20, 1*time.Minute, false, runningNodes)

	c.Stop()
}
