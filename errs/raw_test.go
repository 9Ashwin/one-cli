// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"testing"
)

func TestMarkRaw(t *testing.T) {
	orig := NewValidationError(SubtypeInvalidArgument, "bad")
	raw := MarkRaw(orig)
	if !IsRaw(raw) {
		t.Error("IsRaw(MarkRaw(err)) = false, want true")
	}
	if !errors.Is(raw, orig) {
		t.Error("errors.Is(raw, orig) = false, want true")
	}
	if MarkRaw(nil) != nil {
		t.Error("MarkRaw(nil) must return nil")
	}
}

func TestIsRaw_Negative(t *testing.T) {
	if IsRaw(NewValidationError(SubtypeInvalidArgument, "bad")) {
		t.Error("IsRaw(unmarked typed error) = true, want false")
	}
	if IsRaw(errors.New("plain")) {
		t.Error("IsRaw(plain error) = true, want false")
	}
}
