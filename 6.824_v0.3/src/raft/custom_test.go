package raft

import (
	"bytes"
	"testing"
	"time"

	"../labgob"
)

func persistant_state_to_bytes(ps PersistantState) []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(ps)
	data := w.Bytes()

	return data
}

func TestMake(t *testing.T) {

	// Storing in persister and checking if values stored is retieved correctly
	p_store := MakePersister()
	ps := PersistantState{
		CurrentTerm: 5,
		VotedFor:    4,
		Log:         make(map[int]LogItem),
	}
	ps.Log[7] = LogItem{Term: 4, Entry: ApplyMsg{true, "Please be stored", 7}}
	p_store.SaveRaftState(persistant_state_to_bytes(ps))

	applyCh := make(chan ApplyMsg)

	rf := Make(nil, 0, p_store, applyCh)

	time.Sleep(time.Duration(10) * time.Second)
	ct, isL := rf.GetState()
	if ct != 5 {
		t.Errorf("CurrentTerm not corretly read from persister: Got %d Expceted %d", ct, 5)
	}
	if isL {
		t.Errorf("Node leader just as startup")
	}
	if rf.vs.CommitIndex != 0 {
		t.Errorf("Commit index expceted to start as 0 Got: %d", rf.vs.CommitIndex)
	}

}
