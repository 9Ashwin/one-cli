// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"testing"
)

func TestStaticToken(t *testing.T) {
	tok := Token{AccessToken: "secret", Identity: IdentityUser}
	p := Static{TokenValue: tok}
	got, err := p.Token(context.Background(), IdentityUser)
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got.AccessToken != "secret" {
		t.Errorf("AccessToken = %q, want secret", got.AccessToken)
	}
}

func TestStaticToken_WrongIdentity(t *testing.T) {
	tok := Token{AccessToken: "secret", Identity: IdentityUser}
	p := Static{TokenValue: tok}
	_, err := p.Token(context.Background(), IdentityBot)
	if err == nil {
		t.Fatal("expected error for wrong identity")
	}
}

func TestMapToken(t *testing.T) {
	p := Map{
		IdentityUser: {AccessToken: "u"},
		IdentityBot:  {AccessToken: "b"},
	}
	got, err := p.Token(context.Background(), IdentityBot)
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got.AccessToken != "b" {
		t.Errorf("AccessToken = %q, want b", got.AccessToken)
	}
}
