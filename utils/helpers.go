package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

// FileExists returns true if the specified file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// FindAvailableAPIPort returns the first available port starting from 8080
func FindAvailableAPIPort() int {
	port := 8080
	for {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			return port
		}
		port++
	}
}

// LogDiscussion writes a discussion entry to the chain-specific log file
func LogDiscussion(agentName, message, chainID string, isProposer bool) {
	role := "Validator"
	if isProposer {
		role = "Proposer"
	}

	logEntry := fmt.Sprintf("[%s] %s (%s): %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		agentName,
		role,
		message)

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return
	}

	filename := fmt.Sprintf("logs/discussions_%s.log", chainID)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// ensureDiscussionsDir creates the discussions directory if it doesn't exist
func ensureDiscussionsDir() error {
	if err := os.MkdirAll("data/discussions", 0755); err != nil {
		return fmt.Errorf("failed to create discussions directory: %v", err)
	}
	return nil
}

// GetCurrentRound returns the current round number for the specified chain
func GetCurrentRound(chainID string) int {
	if err := ensureDiscussionsDir(); err != nil {
		log.Printf("Warning: %v", err)
		return 1
	}

	roundFile := fmt.Sprintf("data/discussions/%s_round.txt", chainID)
	data, err := os.ReadFile(roundFile)
	if err != nil {
		if err := os.WriteFile(roundFile, []byte("1"), 0644); err != nil {
			log.Printf("Warning: Failed to create round file: %v", err)
		}
		return 1
	}
	round, _ := strconv.Atoi(string(data))
	return round
}

// IncrementRound increases the round number for the specified chain
func IncrementRound(chainID string) {
	if err := ensureDiscussionsDir(); err != nil {
		log.Printf("Warning: %v", err)
		return
	}

	current := GetCurrentRound(chainID)
	roundFile := fmt.Sprintf("data/discussions/%s_round.txt", chainID)
	if err := os.WriteFile(roundFile, []byte(fmt.Sprintf("%d", current+1)), 0644); err != nil {
		log.Printf("Warning: Failed to increment round: %v", err)
	}
}

// GetDiscussionLog returns the contents of the discussion log for the specified chain
func GetDiscussionLog(chainID string) string {
	if err := ensureDiscussionsDir(); err != nil {
		log.Printf("Warning: %v", err)
		return ""
	}

	logFile := fmt.Sprintf("data/discussions/%s.txt", chainID)
	data, err := os.ReadFile(logFile)
	if err != nil {
		if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
			log.Printf("Warning: Failed to create discussion log: %v", err)
		}
		return ""
	}
	return string(data)
}

// AppendDiscussionLog adds a new message to the discussion log for the specified chain
func AppendDiscussionLog(chainID, message string) {
	if err := ensureDiscussionsDir(); err != nil {
		log.Printf("Warning: %v", err)
		return
	}

	logFile := fmt.Sprintf("data/discussions/%s.txt", chainID)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Failed to open discussion log: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(message + "\n"); err != nil {
		log.Printf("Warning: Failed to append to discussion log: %v", err)
	}
}
