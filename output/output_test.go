// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := JSONFormatter{}
	if err := f.Format(&buf, map[string]any{"ok": true}); err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(buf.String(), `"ok":true`) {
		t.Errorf("output = %q, want ok:true", buf.String())
	}
}

func TestPrettyFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := PrettyFormatter{}
	if err := f.Format(&buf, map[string]any{"ok": true}); err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(buf.String(), "{\n") {
		t.Errorf("output not indented: %q", buf.String())
	}
}
