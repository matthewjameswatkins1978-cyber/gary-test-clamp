package trajectory

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Event represents a single observable event in a worker shift trajectory.
type Event struct {
	Sequence      int    `json:"sequence"`
	Timestamp     string `json:"timestamp"`     // RFC3339
	Type          string `json:"type"`          // e.g. shift_start, shift_end, command_run, file_mutation
	Predecessor   string `json:"predecessor"`   // SHA-256 hash of previous event (or 64 '0's for first event)
	PayloadHash   string `json:"payload_hash"`  // SHA-256 hash of the canonical JSON payload
	Payload       any    `json:"payload"`       // Event-specific payload data
	EventHash     string `json:"event_hash"`    // SHA-256 hash of (predecessor + payload_hash + sequence + type + timestamp)
}

// Writer manages append-only JSONL event writing with SHA-256 hashing.
type Writer struct {
	file         *os.File
	writer       *bufio.Writer
	sequence     int
	lastHash     string
	mu           bool // placeholder for thread safety if needed
}

// NewWriter opens or creates a JSONL trace file for writing.
func NewWriter(path string) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Check if file already has events to resume, or start fresh
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	seq := 0
	lastHash := strings.Repeat("0", 64)

	if info.Size() > 0 {
		// Read existing to find last sequence and hash
		f.Close()
		events, err := ReadEvents(path)
		if err == nil && len(events) > 0 {
			last := events[len(events)-1]
			seq = last.Sequence
			lastHash = last.EventHash
		}
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
	}

	return &Writer{
		file:     f,
		writer:   bufio.NewWriter(f),
		sequence: seq,
		lastHash: lastHash,
	}, nil
}

// ComputePayloadHash computes the SHA-256 hash of the JSON representation of the payload.
func ComputePayloadHash(payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// ComputeEventHash computes the hash for the event record.
func ComputeEventHash(seq int, timestamp, eventType, predecessor, payloadHash string) string {
	data := fmt.Sprintf("%d:%s:%s:%s:%s", seq, timestamp, eventType, predecessor, payloadHash)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// WriteEvent appends a new event of the given type with the given payload.
func (w *Writer) WriteEvent(eventType string, payload any) (*Event, error) {
	w.sequence++
	seq := w.sequence
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	pred := w.lastHash

	payloadHash, err := ComputePayloadHash(payload)
	if err != nil {
		return nil, err
	}

	eventHash := ComputeEventHash(seq, timestamp, eventType, pred, payloadHash)

	ev := &Event{
		Sequence:     seq,
		Timestamp:    timestamp,
		Type:         eventType,
		Predecessor:  pred,
		PayloadHash:  payloadHash,
		Payload:      payload,
		EventHash:    eventHash,
	}

	b, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}

	if _, err := w.writer.Write(append(b, '\n')); err != nil {
		return nil, err
	}

	if err := w.writer.Flush(); err != nil {
		return nil, err
	}

	w.lastHash = eventHash
	return ev, nil
}

// Close closes the underlying file.
func (w *Writer) Close() error {
	if err := w.writer.Flush(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

// ReadEvents reads all events from a JSONL trace file.
func ReadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// VerifyChain verifies the hash chain, sequence continuity, predecessor integrity, and payload hashes of trace events.
func VerifyChain(path string) error {
	events, err := ReadEvents(path)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return errors.New("empty trace file")
	}

	expectedPred := strings.Repeat("0", 64)
	expectedSeq := 1

	for i, ev := range events {
		if ev.Sequence != expectedSeq {
			return fmt.Errorf("sequence break at index %d: expected sequence %d, got %d", i, expectedSeq, ev.Sequence)
		}

		if ev.Predecessor != expectedPred {
			return fmt.Errorf("predecessor mismatch at event sequence %d: expected %s, got %s", ev.Sequence, expectedPred, ev.Predecessor)
		}

		// Verify payload hash
		actualPayloadHash, err := ComputePayloadHash(ev.Payload)
		if err != nil {
			return fmt.Errorf("failed to compute payload hash at event sequence %d: %w", ev.Sequence, err)
		}
		if actualPayloadHash != ev.PayloadHash {
			return fmt.Errorf("payload tampering detected at event sequence %d: expected hash %s, got %s", ev.Sequence, ev.PayloadHash, actualPayloadHash)
		}

		// Verify event hash
		expectedEventHash := ComputeEventHash(ev.Sequence, ev.Timestamp, ev.Type, ev.Predecessor, ev.PayloadHash)
		if expectedEventHash != ev.EventHash {
			return fmt.Errorf("event hash tampering detected at event sequence %d: expected %s, got %s", ev.Sequence, expectedEventHash, ev.EventHash)
		}

		expectedPred = ev.EventHash
		expectedSeq++
	}

	return nil
}
