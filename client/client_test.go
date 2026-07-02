// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package client

import (
	"context"
	"testing"
)

func TestFakeClient(t *testing.T) {
	fake := &Fake{Response: &Response{StatusCode: 200, Body: []byte(`{"ok":true}`)}}
	resp, err := fake.Do(context.Background(), Request{Method: "GET", Path: "/pets"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if len(fake.Requests) != 1 {
		t.Errorf("Requests = %d, want 1", len(fake.Requests))
	}
}
