package communication

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type AgentVote struct {
	ValidatorID   string `json:"validatorId"`
	ValidatorName string `json:"validatorName"`
	Message       string `json:"message"`
	Timestamp     int64  `json:"timestamp"`
	Round         int    `json:"round"`
	Approval      bool   `json:"approval"`
}

var roundRegex = regexp.MustCompile(`\[Round (\d+)\] \((true|false)\) \|@([^|]+)\|: (.+)$`)

// WatchDiscussionFile monitors a discussion file for changes and broadcasts new votes
func WatchDiscussionFile(chainID string, broadcast func(AgentVote)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating file watcher: %v", err)
		return
	}
	defer watcher.Close()

	filename := "data/discussions/mainnet.txt"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("Error creating discussion file: %v", err)
			return
		}
		file.Close()
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading discussion file: %v", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		processLine(line, broadcast)
	}

	if err := watcher.Add(filename); err != nil {
		log.Printf("Error adding file to watcher: %v", err)
		return
	}

	log.Printf("Started watching discussion file: %s", filename)

	lastSize := len(content)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				content, err := os.ReadFile(filename)
				if err != nil {
					log.Printf("Error reading file after change: %v", err)
					continue
				}

				if len(content) > lastSize {
					newContent := string(content[lastSize:])
					lines := strings.Split(newContent, "\n")
					for _, line := range lines {
						if line == "" {
							continue
						}
						processLine(line, broadcast)
					}
					lastSize = len(content)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// processLine parses a line of text into an AgentVote and broadcasts it
func processLine(line string, broadcast func(AgentVote)) {
	matches := roundRegex.FindStringSubmatch(line)

	log.Printf("Line: %s", line)
	log.Printf("All matches: %+v", matches)

	if len(matches) == 5 {
		round := matches[1]
		approval := matches[2] == "true"
		validatorName := matches[3]
		message := strings.TrimSpace(matches[4])

		vote := AgentVote{
			ValidatorID:   validatorName,
			ValidatorName: validatorName,
			Message:       message,
			Timestamp:     time.Now().Unix(),
			Round:         parseInt(round),
			Approval:      approval,
		}

		log.Printf("Broadcasting vote: %+v", vote)
		broadcast(vote)
	} else {
		log.Printf("Line did not match expected format: %s", line)
	}
}

// parseInt converts a string to an integer, returning 0 on error
func parseInt(s string) int {
	val := 0
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		log.Printf("Failed to parse integer from string '%s': %v", s, err)
		return 0
	}
	return val
}
