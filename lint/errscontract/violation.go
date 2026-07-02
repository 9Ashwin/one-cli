// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package errscontract

import "github.com/9Ashwin/one-cli/lint/lintapi"

// Re-export the shared types so existing rule code reads Action /
// Violation locally. The canonical declarations live in lintapi.
type (
	Action    = lintapi.Action
	Violation = lintapi.Violation
)

const (
	ActionReject  = lintapi.ActionReject
	ActionLabel   = lintapi.ActionLabel
	ActionWarning = lintapi.ActionWarning
)
