// Package snowflake provides a very simple Twitter snowflake generator and parser.
package snowflake

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var (
	Epoch int64 = 1288834974657

	NodeBits uint8 = 10

	StepBits uint8 = 12

	nodeMax   int64 = -1 ^ (-1 << NodeBits)
	nodeMask  int64 = nodeMax << StepBits
	stepMask  int64 = -1 ^ (-1 << StepBits)
	timeShift uint8 = NodeBits + StepBits
	nodeShift uint8 = StepBits
)

type Node struct {
	mu   sync.Mutex
	time int64
	node int64
	step int64
}

func NewNode(node int64) (*Node, error) {

	// re-calc in case custom NodeBits or StepBits were set
	nodeMax = -1 ^ (-1 << NodeBits)
	nodeMask = nodeMax << StepBits
	stepMask = -1 ^ (-1 << StepBits)
	timeShift = NodeBits + StepBits
	nodeShift = StepBits

	if node < 0 || node > nodeMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
	}

	return &Node{
		time: 0,
		node: node,
		step: 0,
	}, nil
}

func (n *Node) Generate() int64 {

	n.mu.Lock()

	now := time.Now().UnixNano() / 1000000

	if n.time == now {
		n.step = (n.step + 1) & stepMask

		if n.step == 0 {
			for now <= n.time {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := ((now-Epoch)<<timeShift | (n.node << nodeShift) | (n.step))

	n.mu.Unlock()
	return r
}

func Base36(id int64) string {
	return strconv.FormatInt(id, 36)
}

func Time(id int64) int64 {
	return (id >> timeShift) + Epoch
}

func NodeID(id int64) int64 {
	return id & nodeMask >> nodeShift
}

func Step(id int64) int64 {
	return id & stepMask
}
