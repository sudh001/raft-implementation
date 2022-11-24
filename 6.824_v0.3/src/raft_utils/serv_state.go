package raft_utils


type ServerState struct{
	state int
}

func (this *ServerState) AsFollower(){
	this.state = 0
}

func (this *ServerState) AsCandidate(){
	this.state = 1
}

func (this *ServerState) AsLeader(){
	this.state = 2
}

func (this *ServerState) IsFollower() bool{
	return this.state == 0
}

func (this *ServerState) IsCandidate() bool{
	return this.state == 1
}

func (this *ServerState) IsLeader() bool{
	return this.state == 2
}






