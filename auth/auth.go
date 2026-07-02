// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package auth provides the credential abstraction used by one-cli runtimes.
//
// A CredentialProvider resolves the active identity (user or bot) and returns
// an access token. Implementations may read from OS keychain, config files, or
// environment variables. The interface is intentionally small so generated
// CLIs can plug in platform-specific stores without changing runtime code.
package auth

import (
	"context"
	"fmt"
)

// Identity names the caller persona.
type Identity string

const (
	IdentityUser Identity = "user"
	IdentityBot  Identity = "bot"
)

// Token is a resolved credential.
type Token struct {
	AccessToken string
	Identity    Identity
	Scopes      []string
}

// CredentialProvider resolves tokens for a requested identity.
type CredentialProvider interface {
	// Token returns a token for the requested identity, or an error if no
	// credential is available.
	Token(ctx context.Context, identity Identity) (Token, error)
}

// Static is a CredentialProvider that always returns the same token.
type Static struct {
	TokenValue Token
}

// Token implements CredentialProvider.
func (s Static) Token(_ context.Context, identity Identity) (Token, error) {
	if identity != "" && identity != s.TokenValue.Identity {
		return Token{}, fmt.Errorf("identity %s not available", identity)
	}
	return s.TokenValue, nil
}

// Map is a CredentialProvider backed by a map of identity → token.
type Map map[Identity]Token

// Token implements CredentialProvider.
func (m Map) Token(_ context.Context, identity Identity) (Token, error) {
	if identity == "" {
		identity = IdentityUser
	}
	tok, ok := m[identity]
	if !ok {
		return Token{}, fmt.Errorf("no token for identity %s", identity)
	}
	return tok, nil
}
