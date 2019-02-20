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

// Inserts url into database and returns ID if successful
// Returns error in case of failure
// Duplicate links are automatically ignored
func Push(url string, parent_id int) (int64, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return push(url, parent_id)
}

func push(url string, parent_id int) (int64, error) {
	var id int64 = -1
	if db != nil {
		//stmt, err := db.Prepare("INSERT OR IGNORE INTO queue(link, added_on, status, parent_id) values(?,?,?,?)")
		//if err != nil {
		//	return -1, err
		//}

		res, err := pushStmt.Exec(url, "2017-01-31 10:13", "waiting", parent_id)
		if err != nil {
			return -1, err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return -1, err
		}
	} else {
		log.Fatal("Database not initialized")
	}

	return id, nil
}

func PushMulti(urls []string, parent_id int) ([]int64, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return pushMulti(urls, parent_id)
}

func pushMulti(urls []string, parent_id int) ([]int64, error) {
	var ids []int64
	if db != nil {
		//stmt, err := db.Prepare("INSERT OR IGNORE INTO queue(link, added_on, status, parent_id) values(?,?,?,?)")
		//if err != nil { return ids, err }

		tx, err := db.Begin()
		if err != nil {
			return ids, err
		}

		for _, url := range urls {
			res, err := tx.Stmt(pushMultiStmt).Exec(url, "2017-01-31 10:13", "waiting", parent_id)
			if err != nil {
				return ids, err
			}

			id, err := res.LastInsertId()
			if err != nil {
				return ids, err
			}

			ids = append(ids, id)
		}
		tx.Commit()
	} else {
		log.Fatal("Database not initialized")
	}

	return ids, nil
}

// Returns a node to be crawled, returns error in case fo failure
func Pop() (Node, error) {
	lockDbAccess()
	defer ackDbWorkDone()
	return pop()
}

func pop() (Node, error) {
	if db != nil {
		// Update the row which will be popped

		uuid := pseudo_uuid()
		//sql_update := `UPDATE queue SET status=? WHERE id IN (
		//		    SELECT id from queue WHERE status LIKE "waiting" ORDER BY id ASC LIMIT 1
		//		)`
		//stmt, err := db.Prepare(sql_update)
		//
		//if err != nil {
		//	return Node {}, err
		//} else {
		//	defer stmt.Close();
		//}

		_, err := popStmt.Exec(uuid)

		if err != nil {
			return Node{}, err
		}

		// Pop the row
		sql_select := "SELECT * FROM queue WHERE status LIKE ?"
		rows, err := db.Query(sql_select, uuid)
		if err != nil {
			return Node{}, err
		}

		node := Node{}

		layout := "2006-01-02T15:04:05Z"
		for rows.Next() {
			var crawledOn, addedOn, md5hash sql.NullString
			var matches sql.NullInt64
			err := rows.Scan(&node.Id, &node.Link, &addedOn, &node.Status, &crawledOn, &node.ParentId, &matches, &md5hash)
			if err != nil {
				return Node{}, err
			}

			if crawledOn.Valid {
				node.CrawledOn, err = time.Parse(layout, crawledOn.String)
				if err != nil {
					return Node{}, err
				}
			}
			if addedOn.Valid {
				node.AddedOn, err = time.Parse(layout, addedOn.String)
				if err != nil {
					return Node{}, err
				}
			}
			if matches.Valid {
				node.Matches = matches.Int64
			}
			if md5hash.Valid {
				node.MD5 = md5hash.String
			}
		}

		return node, nil
	} else {
		log.Fatal("Database not initialized")
	}

	return Node{}, nil
}

// CountRemainingRows counts number of remaining rows listed in database
func CountRemainingRows() int {
	count := 0
	sqlCountRemaining := "SELECT COUNT(1) as remaining FROM queue WHERE status LIKE 'waiting'"
	err := db.QueryRow(sqlCountRemaining).Scan(&count)
	if err != nil {
		log.Fatal("Error occuring while reading from database:" + err.Error())
	}
	return count
}

// Update status of row to success or failure
func Update(pk, match_count int64, status, md5hash string) error {
	lockDbAccess()
	defer ackDbWorkDone()
	return update(pk, match_count, status, md5hash)
}

func update(pk, match_count int64, status, md5hash string) error {
	// If status is not success or failure then return error
	if !(status == Success || status == ValidationFailed) {
		return errors.New("Invalid input")
	}

	//sql_update := "UPDATE queue SET status=?, matches = ?, md5 = ? WHERE id = ?"
	//stmt, err := db.Prepare(sql_update)
	//if err != nil {return err}

	_, err := updateStmt.Exec(status, match_count, md5hash, pk)
	if err != nil {
		return err
	}

	return nil
}

func pseudo_uuid() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}
