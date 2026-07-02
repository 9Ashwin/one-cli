// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package runtime

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/9Ashwin/one-cli/auth"
	"github.com/9Ashwin/one-cli/client"
	"github.com/9Ashwin/one-cli/errs"
)

func (r *Runtime) buildAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api METHOD /path",
		Short: "Generic API call",
		Long:  "Make a raw API call with METHOD and path. Use --data for JSON body.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]
			if !strings.HasPrefix(path, "/") {
				return errs.NewValidationError(errs.SubtypeInvalidArgument, "api path must start with /: %s", path).
					WithParam("path")
			}
			data, _ := cmd.Flags().GetString("data")
			req := client.Request{
				Method: method,
				Path:   path,
			}
			if data != "" {
				req.Body = []byte(data)
				if req.Headers == nil {
					req.Headers = map[string]string{}
				}
				req.Headers["Content-Type"] = "application/json"
			}
			return r.execute(cmd.Context(), cmd, req)
		},
	}
	cmd.Flags().String("data", "", "JSON request body")
	return cmd
}

func (r *Runtime) execute(ctx context.Context, cmd *cobra.Command, req client.Request) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	identityStr, _ := cmd.Flags().GetString("as")
	identity := parseIdentity(identityStr)

	if dryRun {
		return r.render(cmd, map[string]any{
			"ok":      true,
			"dry_run": true,
			"request": map[string]any{
				"method":  req.Method,
				"path":    req.Path,
				"query":   req.Query,
				"headers": req.Headers,
				"body":    string(req.Body),
			},
		})
	}

	tok, err := r.deps.Auth.Token(ctx, identity)
	if err != nil {
		return errs.NewAuthenticationError(errs.SubtypeTokenMissing, "no token available: %v", err)
	}
	req.AccessToken = tok.AccessToken

	resp, err := r.deps.Client.Do(ctx, req)
	if err != nil {
		return errs.WrapInternal(err)
	}
	if resp.StatusCode >= 400 {
		return errs.NewAPIError(errs.SubtypeServerError, "API returned status %d", resp.StatusCode).
			WithCode(resp.StatusCode)
	}
	return r.render(cmd, map[string]any{
		"ok":       true,
		"status":   resp.StatusCode,
		"response": jsonString(resp.Body),
	})
}

func jsonString(b []byte) any {
	// Try to decode as JSON object/array; fall back to string.
	var v any
	if err := jsonUnmarshal(b, &v); err == nil {
		return v
	}
	return string(b)
}

func parseIdentity(s string) auth.Identity {
	switch s {
	case "bot":
		return auth.IdentityBot
	case "user", "":
		return auth.IdentityUser
	default:
		return auth.Identity("")
	}
}
