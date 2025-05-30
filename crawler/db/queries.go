package db

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Push(url string, parentID int) (int64, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return push(url, parentID)
}

func push(url string, parentID int) (int64, error) {
	if db == nil {
		return -1, errors.New("database not initialized")
	}

	res, err := pushStmt.Exec(url, time.Now().Format(time.RFC3339), StatusWaiting, parentID)
	if err != nil {
		return -1, fmt.Errorf("failed to insert URL %s: %w", url, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func PushMulti(urls []string, parentID int) ([]int64, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return pushMulti(urls, parentID)
}

func pushMulti(urls []string, parentID int) ([]int64, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var ids []int64
	now := time.Now().Format(time.RFC3339)

	for _, url := range urls {
		res, err := tx.Stmt(pushMultiStmt).Exec(url, now, StatusWaiting, parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert URL %s: %w", url, err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID for URL %s: %w", url, err)
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

func Pop() (Node, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return pop()
}

func pop() (Node, error) {
	if db == nil {
		return Node{}, errors.New("database not initialized")
	}

	uuid := generateUUID()
	_, err := popStmt.Exec(uuid, StatusWaiting)
	if err != nil {
		return Node{}, fmt.Errorf("failed to mark URL as processing: %w", err)
	}

	sqlSelect := "SELECT * FROM queue WHERE status = ?"
	rows, err := db.Query(sqlSelect, uuid)
	if err != nil {
		return Node{}, fmt.Errorf("failed to query for URL: %w", err)
	}
	defer rows.Close()

	var node Node
	for rows.Next() {
		var crawledOn, addedOn, md5hash sql.NullString
		var matches sql.NullInt64

		err := rows.Scan(&node.ID, &node.Link, &addedOn, &node.Status, &crawledOn, &node.ParentID, &matches, &md5hash)
		if err != nil {
			return Node{}, fmt.Errorf("failed to scan row: %w", err)
		}

		if addedOn.Valid {
			node.AddedOn, _ = time.Parse(time.RFC3339, addedOn.String)
		}
		if crawledOn.Valid {
			node.CrawledOn, _ = time.Parse(time.RFC3339, crawledOn.String)
		}
		if matches.Valid {
			node.Matches = matches.Int64
		}
		if md5hash.Valid {
			node.MD5 = md5hash.String
		}
	}

	return node, nil
}

func CountRemainingRows() int {
	var count int
	sqlCountRemaining := "SELECT COUNT(1) FROM queue WHERE status = ?"
	err := db.QueryRow(sqlCountRemaining, StatusWaiting).Scan(&count)
	if err != nil {
		log.Printf("Error counting remaining rows: %v", err)
		return 0
	}
	return count
}

func Update(pk, matchCount int64, status, md5hash string) error {
	lockDbAccess()
	defer ackDbWorkDone()
	return update(pk, matchCount, status, md5hash)
}

func update(pk, matchCount int64, status, md5hash string) error {
	if status != StatusSuccess && status != StatusValidationFailed {
		return fmt.Errorf("invalid status '%s': must be '%s' or '%s'", 
			status, StatusSuccess, StatusValidationFailed)
	}

	_, err := updateStmt.Exec(status, matchCount, md5hash, pk)
	if err != nil {
		return fmt.Errorf("failed to update row %d: %w", pk, err)
	}

	return nil
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Error generating UUID: %v", err)
		return ""
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
