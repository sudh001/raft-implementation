package raft

//import "../raft_utils"
import "fmt"
import "sync"


type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

func (rf *Raft) startElection() {
	
	fmt.Println("Starting election", rf.me)
	rf.mu.Lock()
	defer rf.restart_election_timer()
	
	
	if !rf.ss.IsFollower() || rf.killed(){
		rf.mu.Unlock()
		return
	}
	// am a followr, need to become candidate
	// increment my term, vote for myself and send reqs to others
	rf.ps.CurrentTerm += 1
	rf.ss.AsCandidate()
	rf.ps.VotedFor = rf.me
	rf.persist() // dosnt lock
	
	Votes := 1
	var Vote_mu sync.Mutex
	// preparing args for others
	
	args := RequestVoteArgs{
		Term: rf.ps.CurrentTerm,
		CandidateId: rf.me,
		LastLogIndex: -1,
		LastLogTerm: 0,
	}
	fmt.Printf("%d to send req votes with term: %d\n",rf.me, rf.ps.CurrentTerm)
	rf.mu.Unlock()

	// sending requests out
	
	
	for s_id, _ := range rf.peers{
		if s_id != rf.me{
			peer_id := s_id
			fmt.Println(peer_id, rf.me)
			go func(){
				var reply RequestVoteReply
				fmt.Printf("sending req from %d to %d\n", rf.me, peer_id)
				ok := rf.peers[peer_id].Call("Raft.RequestVote", &args, &reply)
				rf.mu.Lock()
				defer rf.mu.Unlock()
				// am i still candidate with the appropriate term or did i evolve ?
				if !ok || !(rf.ss.IsCandidate() && rf.ps.CurrentTerm == args.Term){
					
					return
				}
				fmt.Println("Got reply from: ", peer_id, "to" , rf.me, "contents: ", reply)
				// the reply got is important for me to read as i didnt evolve
				if reply.VoteGranted{
					fmt.Println("Got vote", rf.me)
					Vote_mu.Lock()
					Votes += 1
					
					
					// seeing if i have what it takes to become leader
					if Votes >= int(len(rf.peers)/2) + 1{
						// I do!
						fmt.Println("I am Leader!!", rf.me)
						rf.ss.AsLeader()
						rf.leaderStateInit()
						rf.election_timer.Stop()
						rf.ls.HeartbeatTimer.Start()
						
					}
					
					Vote_mu.Unlock()
					
				} else if reply.Term > rf.ps.CurrentTerm{
					// need to become follower
					rf.ps.CurrentTerm = reply.Term
					rf.ps.VotedFor = -1
					rf.persist()
					rf.ss.AsFollower()
					rf.election_timer.Stop()
					rf.restart_election_timer()
				}
 				
			
			}()
		
		}
	}
	
	
	
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	rf.mu.Lock()
	defer rf.persist()
	defer rf.mu.Unlock()

	if args.Term < rf.ps.CurrentTerm {
		reply.Term = rf.ps.CurrentTerm
		reply.VoteGranted = false
		return
	}

	if args.Term > rf.ps.CurrentTerm || (args.Term == rf.ps.CurrentTerm && rf.ps.VotedFor == -1) {
		LLI := rf.GetLastIndex()
		LLT := rf.GetLastTerm(LLI)

		if LLT < args.LastLogTerm || (LLT == args.LastLogTerm && LLI <= args.LastLogIndex) {
			rf.ps.CurrentTerm = args.Term
			rf.ps.VotedFor = args.CandidateId
			rf.ss.AsFollower()

			reply.Term = rf.ps.CurrentTerm
			reply.VoteGranted = true
			
			// need to restart timer
			// case 2 of restart election timer
			rf.restart_election_timer()

			return
		}
	}

	reply.Term = rf.ps.CurrentTerm
	reply.VoteGranted = false

	//send RPC??
}

