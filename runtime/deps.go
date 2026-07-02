// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package runtime

import (
	"github.com/9Ashwin/one-cli/auth"
	"github.com/9Ashwin/one-cli/client"
	"github.com/9Ashwin/one-cli/metadata"
	"github.com/9Ashwin/one-cli/output"
	"github.com/9Ashwin/one-cli/vfs"
)

// Deps are the runtime dependencies injected by the generated CLI.
type Deps struct {
	Meta   *metadata.Spec
	Auth   auth.CredentialProvider
	Client client.APIClient
	Output output.Formatter
	VFS    vfs.VFS
}
