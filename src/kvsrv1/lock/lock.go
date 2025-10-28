package lock

import (
	"time"

	"6.5840/kvsrv1/rpc"
	kvtest "6.5840/kvtest1"
)

// distributed lock, using key/value clerk, 相当于lockKey/clntID作为key/value
type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk
	// You may add code here
	clntID  string // a unique identifier for each lock client
	lockKey string
}

// The tester calls MakeLock() and passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
//
// Use l as the key to store the "lock state" (you would have to decide
// precisely what the lock state is).
func MakeLock(ck kvtest.IKVClerk, l string) *Lock {
	lk := &Lock{ck: ck}
	// You may add code here
	lk.clntID = kvtest.RandValue(8)
	lk.lockKey = l
	return lk
}

func (lk *Lock) Acquire() {
	// Your code here
	for {
		value, version, err := lk.ck.Get(lk.lockKey)
		if err == rpc.ErrNoKey || (err == rpc.OK && value == "") {
			err = lk.ck.Put(lk.lockKey, lk.clntID, version)

			if err == rpc.OK {
				return
			}
		} else if err == rpc.OK && value == lk.clntID {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (lk *Lock) Release() {
	// Your code here
	for {
		value, version, err := lk.ck.Get(lk.lockKey)
		if err == rpc.OK && value == lk.clntID {
			err = lk.ck.Put(lk.lockKey, "", version)

			if err == rpc.OK {
				return
			}
		} else {
			return
		}
	}
}
