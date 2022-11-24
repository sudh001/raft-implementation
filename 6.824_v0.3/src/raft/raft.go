package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"sync"
	"sync/atomic"

	"../labgob"
	"../labrpc"
	"../raft_utils"
)

const ELE_ST, ELE_EN = 1000, 1200
const HB_T = 100


type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type LogItem struct {
	Entry ApplyMsg
	Term  int
}

//
// A Go object implementing a single Raft peer.
//

type PersistantState struct {
	CurrentTerm int
	VotedFor    int
	Log         map[int]LogItem
}

type VolatileState struct {
	CommitIndex int
	LastApplied int
}

type LeaderState struct {
	NextIndex  map[int]int
	MatchIndex map[int]int
	//TODO: Add heartbeat timer
	HeartbeatTimer raft_utils.Timer
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	ps PersistantState
	vs VolatileState
	ls *LeaderState // will hold nil if server isnt leader
	ss raft_utils.ServerState
	//TODO: Add election timer

	election_timer raft_utils.Timer
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.ps.CurrentTerm
	isleader = rf.ss.IsLeader()

	return term, isleader
}

func (rf *Raft) GetLastIndex() int {
	if len(rf.ps.Log) < 1 {
		return -1
	}

	keys := make([]int, 0, len(rf.ps.Log))

	for k := range rf.ps.Log {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys[0]
}

func (rf *Raft) GetLastTerm(index int) int {
	if len(rf.ps.Log) < 1 || index == -1 {
		return 0
	}
	return rf.ps.Log[index].Term
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
// persist will not obtain lock.. it is the duty of the caller to obtain lock before calling persist
func (rf *Raft) persist() {
	
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	
	e.Encode(rf.ps)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)

}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?

		ps_temp := PersistantState{
			CurrentTerm: 0,
			VotedFor:    -1,
			Log:         make(map[int]LogItem),
		}
		rf.mu.Lock()
		defer rf.mu.Unlock()

		rf.ps = ps_temp
		return
	}

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var ps_temp PersistantState

	if err := d.Decode(&ps_temp); err != nil {
		log.Fatal(err)
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.ps = ps_temp

}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//


func (rf *Raft) isLogUpToDate(LastIndex int, LastTerm int) bool {
	RaftLastIndex := rf.GetLastIndex()
	RaftLastTerm := rf.GetLastTerm(RaftLastIndex)

	if LastTerm == RaftLastTerm {
		return LastIndex >= RaftLastIndex
	}

	return LastTerm > RaftLastTerm
}

// to restart election timer.. obtain lock before calling 

func (rf *Raft)restart_election_timer(){

	if !(rf.killed()){
		rf.election_timer = *raft_utils.MakeTimer(ELE_ST, ELE_EN, rf.startElection)
		rf.election_timer.Start()
	}

}




func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
	// TODO: Stop running timers
	rf.election_timer.Stop()
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	fmt.Println("num servers", len(peers))
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.vs = VolatileState{CommitIndex: 0, LastApplied: 0}
	rf.ls = nil
	rf.ss.AsFollower()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// setting Log[-1]: term 0, no applyMsg in the index (not to be sent back to client)
	rf.ps.Log[-1] = LogItem{ApplyMsg{}, 0}
	rf.persist()

	//TODO: Start election timer

	rf.restart_election_timer()

	return rf
}
