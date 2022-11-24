package raft

import (
	"../raft_utils"
	"fmt"
)

type AppendEntry struct {
	Term              int
	LeaderId          int
	PrevLogIndex      int
	PrevLogTerm       int
	Entries           map[int]ApplyMsg
	LeaderCommitIndex int
}

type AppendEntryReply struct {
	Term    int
	Success bool
}

func (rf *Raft) leaderStateInit() {
	nextInd := make(map[int]int)
	nextInd[rf.me] = 0

	matchInd := make(map[int]int)
	matchInd[rf.me] = -1
	rf.ls = &LeaderState{}
	rf.ls.NextIndex = nextInd
	rf.ls.MatchIndex = matchInd
	rf.ls.HeartbeatTimer = *raft_utils.MakeTimer(HB_T, HB_T + 1, rf.SendHeartBeat)

}

func (rf *Raft) SendHeartBeat() {

	// a heartbeat is just an append entry with no log entries and sould be treated as such
	fmt.Println("Sending heartbeat", rf.me)
	for peer_id, _ := range rf.peers {
		if peer_id != rf.me {
			go func() {
				rf.mu.Lock()
				// if not leader, do nothing
				if !rf.ss.IsLeader() {
					rf.mu.Unlock()
					return
				}
				// creating append entry
				hb := AppendEntry{
					Term:         rf.ps.CurrentTerm,
					LeaderId:     rf.me,
					PrevLogIndex: -1, // need to change this to a general value
					PrevLogTerm:  0,  // need to change this to a general value
					// not worrying about Entries
					LeaderCommitIndex: -1, // need to change this to a general value

				}

				var reply AppendEntryReply
				rf.mu.Unlock()

				// a waiting call
				if ok := rf.peers[peer_id].Call("Raft.RcvAppendEntry", &hb, &reply); !ok {
					return
				}

				// parsing the reply
				rf.mu.Lock()
				// if not leader do nothing
				if !rf.ss.IsLeader() {
					rf.mu.Unlock()
					return
				}

				// if i see a term greater than mine, i shall step down
				if !reply.Success && reply.Term > rf.ps.CurrentTerm {
					rf.ls.HeartbeatTimer.Stop()
					rf.ls = nil
					rf.ss.AsFollower()
					rf.ps.CurrentTerm = reply.Term
					rf.ps.VotedFor = -1
					rf.persist()

					// TODO: start election timer
					rf.election_timer.Stop()
					rf.restart_election_timer()
				}
				rf.mu.Unlock()
			}()
		}

	}
	// to restart heartbeat timer
	rf.mu.Lock()
	if rf.ss.IsLeader(){
		rf.ls.HeartbeatTimer = *raft_utils.MakeTimer(HB_T, HB_T + 1, rf.SendHeartBeat)
		rf.ls.HeartbeatTimer.Start()
	
	}
	rf.mu.Unlock()

}

func (rf *Raft) RcvAppendEntry(ae *AppendEntry, ae_r *AppendEntryReply) {

	rf.mu.Lock()

	var success bool
	if ae.Term < rf.ps.CurrentTerm {
		success = false
	} else if logItem, ok := rf.ps.Log[ae.PrevLogIndex]; !ok || logItem.Term != ae.PrevLogTerm {
		success = false
	} else {
		// case 3 replying to current leader
		rf.election_timer.Stop()
		rf.restart_election_timer()
		success = true
	}

	rf.mu.Unlock()

	// not worrying about log catchup as log replication not considered now

	reply := AppendEntryReply{
		Term:    rf.ps.CurrentTerm,
		Success: success,
	}

	*ae_r = reply

}
