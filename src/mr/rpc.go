package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type TaskType int

const (
	MapTask TaskType = iota
	ReduceTask
	Wait
	Complete
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type TaskReqArgs struct {
}

type TaskReqReply struct {
	Type     TaskType
	TaskID   int
	FileName string
	NMap     int
	NReduce  int
}

type TaskDoneArgs struct {
	Type   TaskType
	TaskID int
	// OutputFileName string
}

type TaskDoneReply struct {
	Ok bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
