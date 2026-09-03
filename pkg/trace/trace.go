package trace

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Data      any       `json:"data"`
	PrevHash  string    `json:"prev_hash"`
	Hash      string    `json:"hash"`
}

type Logger struct {
	path     string
	prevHash string
}

func NewLogger(path string) (*Logger, error) {
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			var lastLine string
			for scanner.Scan() {
				lastLine = scanner.Text()
			}
			if lastLine != "" {
				var ev Event
				if json.Unmarshal([]byte(lastLine), &ev) == nil && ev.Hash != "" {
					prevHash = ev.Hash
				}
			}
		}
	}
	return &Logger{path: path, prevHash: prevHash}, nil
}

func (l *Logger) Log(eventType string, data any) (Event, error) {
	ev := Event{
		Timestamp: time.Now().UTC(),
		Type:      eventType,
		Data:      data,
		PrevHash:  l.prevHash,
	}
	
	// Compute hash excluding Hash field
	hashContent, err := json.Marshal(struct {
		Timestamp time.Time `json:"timestamp"`
		Type      string    `json:"type"`
		Data      any       `json:"data"`
		PrevHash  string    `json:"prev_hash"`
	}{
		Timestamp: ev.Timestamp,
		Type:      ev.Type,
		Data:      ev.Data,
		PrevHash:  ev.PrevHash,
	})
	if err != nil {
		return Event{}, err
	}

	sum := sha256.Sum256(hashContent)
	ev.Hash = hex.EncodeToString(sum[:])

	lineBytes, err := json.Marshal(ev)
	if err != nil {
		return Event{}, err
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return Event{}, err
	}
	defer f.Close()

	if _, err := f.WriteString(string(lineBytes) + "\n"); err != nil {
		return Event{}, err
	}

	l.prevHash = ev.Hash
	return ev, nil
}

func Verify(path string) (bool, int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, 0, nil
		}
		return false, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	expectedPrev := "0000000000000000000000000000000000000000000000000000000000000000"
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return false, count, fmt.Errorf("malformed JSON at line %d: %w", count+1, err)
		}

		if ev.PrevHash != expectedPrev {
			return false, count, fmt.Errorf("prev_hash mismatch at line %d: got %s, want %s", count+1, ev.PrevHash, expectedPrev)
		}

		// Recompute hash
		hashContent, err := json.Marshal(struct {
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
			Data      any       `json:"data"`
			PrevHash  string    `json:"prev_hash"`
		}{
			Timestamp: ev.Timestamp,
			Type:      ev.Type,
			Data:      ev.Data,
			PrevHash:  ev.PrevHash,
		})
		if err != nil {
			return false, count, err
		}

		sum := sha256.Sum256(hashContent)
		computedHash := hex.EncodeToString(sum[:])

		if ev.Hash != computedHash {
			return false, count, fmt.Errorf("hash tamper detected at line %d: stored %s, computed %s", count+1, ev.Hash, computedHash)
		}

		expectedPrev = ev.Hash
		count++
	}

	return true, count, nil
}
