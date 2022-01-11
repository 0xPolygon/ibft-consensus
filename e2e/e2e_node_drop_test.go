package e2e

import (
	"testing"
	"time"
)

func TestE2E_NodeDrop(t *testing.T) {
	c := newPBFTCluster(t, "node_drop", "ptr", 5)
	c.Start()

	// wait for two heights and stop node 1
	err := c.WaitForHeight(2, 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	c.StopNode("ptr_0")
	err = c.WaitForHeight(15, 1*time.Minute, generateNodeNames(1, 4, "ptr_"))
	if err != nil {
		t.Fatal(err)
	}
}
