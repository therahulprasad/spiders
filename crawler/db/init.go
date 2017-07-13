package db

import "database/sql"

var db *sql.DB = nil
var chWait, chWaitAck chan bool
var pushStmt, pushMultiStmt, popStmt, updateStmt *sql.Stmt

// This will be executed only once
func init() {
	// Instantiate global variables
	chWait = make(chan bool)
	chWaitAck = make(chan bool)

	// Lets only single instance access DB preventing Locking errors
	// TODO: Find out a better way to this
	go multiplexer()
}

// This is run as goroutine, every function making DB calls first locks the db and then access DB
func multiplexer() {
	for {
		<-chWait
		<-chWaitAck
	}
}

// Multiplexer is running single loop so only one instance can access DB at a time
func lockDbAccess() {
	chWait <- true
}

func ackDbWorkDone() {
	chWaitAck<- true
}
