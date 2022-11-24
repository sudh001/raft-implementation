package raft_utils


import ( 
	"testing"
	"sync"
	"time"
)

type state struct{
	mu sync.Mutex
	val int
}

func (this *state)Update(){
	this.mu.Lock()
	this.val += 1
	this.mu.Unlock()
}
func (this *state) get_val() int{
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.val
}

func TestTimer(tester *testing.T){

	// Tests whether timer calls a func some number of times within a predictable interval and stops 

	st := 1
	en := 2
	s := state{}
	t := MakeTimer(st, en, s.Update)
	
	// 
	go func(){
		for {
			if s.get_val() == 10{
				t.Stop()
				return
			}
		}

	}()
	t.Start()
	time.Sleep(1000 * time.Millisecond)
	if v := s.get_val(); v != 10{
		tester.Errorf("timer not stopping properly, Expected state_val: %d Got: %d", 10,v)
	}
	
}



