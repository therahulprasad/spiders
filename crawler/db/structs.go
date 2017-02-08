package db

import "time"

type Node struct {
	Id int64
	Link string
	AddedOn time.Time
	Status string
	CrawledOn time.Time
	ParentId int
	Matches int64
}

const (
	ValidationFailed = "ValidationFailed"
	Success = "Success"
)