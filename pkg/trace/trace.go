package trace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"telltail/pkg/models"
)

// CalculateHash computes the SHA-256 hash of an event combined with the previous hash
func CalculateHash(seq int, timestamp string, eventType string, payload map[string]interface{}, prevHash string) string {
	payloadBytes, err := json.Marshal(payload)
	payloadStr := "{}"
	if err == nil && len(payloadBytes) > 0 && string(payloadBytes) != "null" {
		payloadStr = string(payloadBytes)
	}

	data := fmt.Sprintf("%d|%s|%s|%s|%s", seq, timestamp, eventType, payloadStr, prevHash)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// LoadTrace reads and parses a JSONL trace file
func LoadTrace(filePath string) ([]models.Event, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Event{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var events []models.Event
	decoder := json.NewDecoder(file)
	for {
		var ev models.Event
		if err := decoder.Decode(&ev); err == io.EOF {
			break
		} else if err != nil {
			// Fail-closed behaviour on malformed trace line
			return nil, fmt.Errorf("malformed trace line: %w", err)
		}
		events = append(events, ev)
	}
	return events, nil
}

// VerifyTrace checks the contiguous sequence indices and SHA-256 hash chain
func VerifyTrace(events []models.Event) (bool, error) {
	if len(events) == 0 {
		return true, nil
	}

	prevHash := ""
	for i, ev := range events {
		if ev.Seq != i {
			return false, fmt.Errorf("sequence break: expected seq %d, got %d", i, ev.Seq)
		}

		expectedHash := CalculateHash(ev.Seq, ev.Timestamp, ev.EventType, ev.Payload, prevHash)
		if ev.Hash != expectedHash {
			return false, fmt.Errorf("tamper detected at seq %d: expected hash %s, got %s", ev.Seq, expectedHash, ev.Hash)
		}
		prevHash = ev.Hash
	}

	return true, nil
}

// AppendEvent appends a new event to the trace file, calculating its hash relative to the previous event in the file
func AppendEvent(filePath string, eventType string, payload map[string]interface{}) (models.Event, error) {
	events, err := LoadTrace(filePath)
	if err != nil {
		// If load fails because of malformed lines, we fail closed
		return models.Event{}, fmt.Errorf("failed to load trace prior to append: %w", err)
	}

	seq := len(events)
	prevHash := ""
	if seq > 0 {
		prevHash = events[seq-1].Hash
	}

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	hash := CalculateHash(seq, timestamp, eventType, payload, prevHash)

	ev := models.Event{
		Seq:       seq,
		Timestamp: timestamp,
		EventType: eventType,
		Payload:   payload,
		Hash:      hash,
	}

	line, err := json.Marshal(ev)
	if err != nil {
		return models.Event{}, err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return models.Event{}, err
	}
	defer file.Close()

	if _, err := file.Write(append(line, '\n')); err != nil {
		return models.Event{}, err
	}

	return ev, nil
}

// VerifyTraceFile verifies a trace file directly by loading and verifying
func VerifyTraceFile(filePath string) (bool, error) {
	events, err := LoadTrace(filePath)
	if err != nil {
		return false, err
	}
	if len(events) == 0 {
		return false, fmt.Errorf("empty trace file or does not exist")
	}
	return VerifyTrace(events)
}

// SafePayload ensures that secrets aren't logged in traces. We replace known sensitive values.
func SafePayload(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return make(map[string]interface{})
	}
	clean := make(map[string]interface{})
	for k, v := range payload {
		// Redact standard secret fields
		kLower := strings.ToLower(k)
		if strings.Contains(kLower, "key") || strings.Contains(kLower, "secret") || strings.Contains(kLower, "password") || strings.Contains(kLower, "token") {
			clean[k] = "[REDACTED]"
		} else {
			clean[k] = v
		}
	}
	return clean
}
