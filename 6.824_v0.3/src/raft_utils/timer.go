package raft_utils

import (

	

	"time"
	"math/rand"
)

type TimeoutHandler func()

type Timer struct {

	TimeoutStart int
	TimeoutEnd   int
	TimeoutValue int
	Th           TimeoutHandler
	IsAlive      bool
}

func MakeTimer(start int, end int, th TimeoutHandler) *Timer{
	t := &Timer{}
	t.Init(start, end, th)
	return t
}

func (t *Timer) Init(start int, end int, th TimeoutHandler) {
	

	t.TimeoutStart = start
	t.TimeoutEnd = end
	t.Th = th
	t.IsAlive = true
}

func (t *Timer) reset() {
	

	t.TimeoutValue = rand.Intn(t.TimeoutEnd-t.TimeoutStart) + t.TimeoutStart

	
	time.Sleep(time.Duration(t.TimeoutValue) * time.Millisecond)
	
	

	t.OnTimeout()
}

func (t *Timer) Start() {
	
	go t.reset()
	//time.Sleep(300 * time.Second)
}

func (t *Timer) OnTimeout() {


	if t.IsAlive == true {
		
		t.Th() // handler expected to be non-waiting

	}
}

func (t *Timer) Stop() {
	//t.stop => (isALive =false)

	t.IsAlive = false
	
}

