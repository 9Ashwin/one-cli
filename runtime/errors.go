// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package runtime

import (
	"github.com/9Ashwin/one-cli/errs"
)

func errInternal(msg string) error {
	return errs.NewInternalError(errs.SubtypeSDKError, msg)
}

func errValidation(subtype errs.Subtype, msg string) error {
	return errs.NewValidationError(subtype, msg)
}
