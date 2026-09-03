package trajectory

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Event represents a single observable event in a shift trajectory.
type Event struct {
	Seq       int64                  `json:"seq"`
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	PrevHash  string                 `json:"prev_hash"`
	Hash      string                 `json:"hash"`
}

// Writer manages writing hash-chained JSONL events to a file.
type Writer struct {
	file     *os.File
	path     string
	seq      int64
	prevHash string
}

// NewWriter creates or appends to a trajectory trace file.
func NewWriter(path string) (*Writer, error) {
	// Check if file exists and read last event to continue hash chain if appending,
	// or create a new file. For simplicity, we open with create/append.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	var seq int64 = 0
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"

	// Read existing lines to find last seq and hash if file is non-empty
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err == nil {
			seq = ev.Seq
			prevHash = ev.Hash
		}
	}

	// Seek to end for appending
	_, err = f.Seek(0, io.SeekEnd)
	if err != nil {
		f.Close()
		return nil, err
	}

	return &Writer{
		file:     f,
		path:     path,
		seq:      seq,
		prevHash: prevHash,
	}, nil
}

// WriteEvent records a new event into the trajectory log with SHA-256 hash chaining.
func (w *Writer) WriteEvent(eventType string, payload map[string]interface{}) (*Event, error) {
	w.seq++
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if payload == nil {
		payload = make(map[string]interface{})
	}

	ev := Event{
		Seq:       w.seq,
		Timestamp: now,
		EventType: eventType,
		Payload:   payload,
		PrevHash:  w.prevHash,
	}

	// Calculate hash over canonical fields: seq, timestamp, event_type, payload, prev_hash
	hBytes, err := computeHash(ev.Seq, ev.Timestamp, ev.EventType, ev.Payload, ev.PrevHash)
	if err != nil {
		return nil, err
	}
	ev.Hash = hBytes

	line, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}

	_, err = w.file.Write(append(line, '\n'))
	if err != nil {
		return nil, err
	}

	w.prevHash = ev.Hash
	return &ev, nil
}

// Close closes the writer file handle.
func (w *Writer) Close() error {
	return w.file.Close()
}

func computeHash(seq int64, timestamp, eventType string, payload map[string]interface{}, prevHash string) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s:%s:%s", seq, timestamp, eventType, string(payloadBytes), prevHash)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Verify validates a trajectory file for sequence gaps, predecessor mismatches, and content tampering.
func Verify(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var expectedSeq int64 = 0
	expectedPrevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
		}

		expectedSeq++
		if ev.Seq != expectedSeq {
			return fmt.Errorf("line %d: sequence break: expected seq %d, got %d", lineNum, expectedSeq, ev.Seq)
		}

		if ev.PrevHash != expectedPrevHash {
			return fmt.Errorf("line %d: predecessor hash mismatch: expected %s, got %s", lineNum, expectedPrevHash, ev.PrevHash)
		}

		recomputed, err := computeHash(ev.Seq, ev.Timestamp, ev.EventType, ev.Payload, ev.PrevHash)
		if err != nil {
			return fmt.Errorf("line %d: hash computation error: %w", lineNum, err)
		}

		if ev.Hash != recomputed {
			return fmt.Errorf("line %d: content tampering detected: expected hash %s, got %s", lineNum, recomputed, ev.Hash)
		}

		expectedPrevHash = ev.Hash
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if expectedSeq == 0 {
		return fmt.Errorf("trajectory file is empty")
	}

	return nil
}
