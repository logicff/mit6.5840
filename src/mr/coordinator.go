package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	Task_Timeout = 10 // the time of task timeout (second)
)

type Coordinator struct {
	// Your definitions here.
	nMap        int
	nReduce     int
	files       []string
	nMapDone    int
	nReduceDone int
	mapTasks    map[int]*TaskStatus
	reduceTasks map[int]*TaskStatus
	mu          sync.Mutex
}

type TaskStatus struct {
	Status    int // 0: not allocated, 1: processing, 2: finished
	StartTime time.Time
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) AllocateTask(args *TaskReqArgs, reply *TaskReqReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.nMapDone < c.nMap {
		for taskID, status := range c.mapTasks {
			if status.Status == 0 {
				reply.Type = MapTask
				reply.TaskID = taskID
				reply.NReduce = c.nReduce
				reply.FileName = c.files[taskID]
				c.mapTasks[taskID].Status = 1
				c.mapTasks[taskID].StartTime = time.Now()
				return nil
			}
		}
		reply.Type = Wait
	} else if c.nReduceDone < c.nReduce {
		for taskID, status := range c.reduceTasks {
			if status.Status == 0 {
				reply.Type = ReduceTask
				reply.TaskID = taskID
				reply.NMap = c.nMap
				c.reduceTasks[taskID].Status = 1
				c.reduceTasks[taskID].StartTime = time.Now()
				return nil
			}
		}
		reply.Type = Wait
	} else {
		reply.Type = Complete
	}
	return nil
}

func (c *Coordinator) HandleTaskFinish(args *TaskDoneArgs, reply *TaskDoneReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch args.Type {
	case MapTask:
		if _, ok := c.mapTasks[args.TaskID]; ok {
			c.nMapDone++
			c.mapTasks[args.TaskID].Status = 2
			reply.Ok = true
		}
	case ReduceTask:
		if _, ok := c.reduceTasks[args.TaskID]; ok {
			c.nReduceDone++
			c.reduceTasks[args.TaskID].Status = 2
			reply.Ok = true
		}
	default:
		reply.Ok = false
	}
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// handle the timeout task
func (c *Coordinator) monitorTimeoutTasks() {
	for {
		time.Sleep(1 * time.Second)
		c.mu.Lock()
		now := time.Now()
		if c.nMapDone < c.nMap {
			for taskID, status := range c.mapTasks {
				if status.Status == 1 {
					if now.Sub(status.StartTime) > Task_Timeout*time.Second {
						c.mapTasks[taskID].Status = 0
					}
				}
			}
		}
		if c.nReduceDone < c.nReduce {
			for taskID, status := range c.reduceTasks {
				if status.Status == 1 {
					if now.Sub(status.StartTime) > Task_Timeout*time.Second {
						c.reduceTasks[taskID].Status = 0
					}
				}
			}
		}
		c.mu.Unlock()

		if c.Done() {
			return
		}
	}
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	// ret := false

	// Your code here.
	ret := c.nReduceDone == c.nReduce

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.files = files
	c.nMap = len(files)
	c.nReduce = nReduce
	c.nMapDone = 0
	c.nReduceDone = 0
	c.mapTasks = make(map[int]*TaskStatus)
	c.reduceTasks = make(map[int]*TaskStatus)

	for i := 0; i < c.nMap; i++ {
		c.mapTasks[i] = &TaskStatus{
			Status: 0,
		}
	}

	for i := 0; i < c.nReduce; i++ {
		c.reduceTasks[i] = &TaskStatus{
			Status: 0,
		}
	}

	go c.monitorTimeoutTasks()

	c.server()
	return &c
}
