package clamp

import "testing"

func TestClampBelowMin(t *testing.T) {
	if got := Clamp(-5, 0, 10); got != 0 {
		t.Errorf("Clamp(-5, 0, 10) = %d, want 0", got)
	}
}

func TestClampAboveMax(t *testing.T) {
	if got := Clamp(15, 0, 10); got != 10 {
		t.Errorf("Clamp(15, 0, 10) = %d, want 10", got)
	}
}

func TestClampInsideRange(t *testing.T) {
	if got := Clamp(5, 0, 10); got != 5 {
		t.Errorf("Clamp(5, 0, 10) = %d, want 5", got)
	}
}

func TestClampAtMin(t *testing.T) {
	if got := Clamp(0, 0, 10); got != 0 {
		t.Errorf("Clamp(0, 0, 10) = %d, want 0", got)
	}
}

func TestClampAtMax(t *testing.T) {
	if got := Clamp(10, 0, 10); got != 10 {
		t.Errorf("Clamp(10, 0, 10) = %d, want 10", got)
	}
}
