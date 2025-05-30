package db

import "time"

type Node struct {
	ID        int64
	Link      string
	AddedOn   time.Time
	Status    string
	CrawledOn time.Time
	ParentID  int
	Matches   int64
	MD5       string
}

const (
	StatusWaiting           = "waiting"
	StatusProcessing        = "processing"
	StatusSuccess           = "success"
	StatusValidationFailed  = "validation_failed"
)

// Legacy constants for backward compatibility
const (
	ValidationFailed = StatusValidationFailed
	Success          = StatusSuccess
)