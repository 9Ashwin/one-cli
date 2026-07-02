// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package runtime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/9Ashwin/one-cli/client"
	"github.com/9Ashwin/one-cli/errs"
	"github.com/9Ashwin/one-cli/metadata"
)

func (r *Runtime) buildServiceCmd(svc metadata.Service) *cobra.Command {
	svcCmd := &cobra.Command{
		Use:   svc.Name,
		Short: "Commands for " + svc.Name,
	}
	for _, res := range sortedResources(svc.Resources) {
		svcCmd.AddCommand(r.buildResourceCmd(svc, res))
	}
	// Attach shortcut commands directly to the service command.
	for _, shortcut := range r.serviceShortcuts(svc) {
		svcCmd.AddCommand(r.buildShortcutCmd(svc, shortcut))
	}
	return svcCmd
}

func (r *Runtime) buildResourceCmd(svc metadata.Service, res metadata.Resource) *cobra.Command {
	resCmd := &cobra.Command{
		Use:   res.Name,
		Short: "Commands for " + svc.Name + "." + res.Name,
	}
	for _, m := range res.MethodList() {
		resCmd.AddCommand(r.buildMethodCmd(svc, res, m))
	}
	return resCmd
}

func (r *Runtime) buildMethodCmd(svc metadata.Service, res metadata.Resource, m metadata.Method) *cobra.Command {
	cmd := &cobra.Command{
		Use:   m.OperationID,
		Short: m.Summary,
		Long:  m.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := r.buildRequest(cmd, m)
			if err != nil {
				return err
			}
			return r.execute(cmd.Context(), cmd, req)
		},
	}
	for _, p := range m.Params {
		addParamFlag(cmd, p)
	}
	return cmd
}

func (r *Runtime) buildRequest(cmd *cobra.Command, m metadata.Method) (client.Request, error) {
	req := client.Request{
		Method: m.HTTPMethod,
		Path:   m.Path,
		Query:  map[string]string{},
	}
	for _, p := range m.Params {
		flagName := paramFlagName(p.Name)
		if !cmd.Flags().Changed(flagName) && p.Required {
			return client.Request{}, errs.NewValidationError(errs.SubtypeInvalidArgument, "required flag --%s is missing", flagName).
				WithParam("--" + flagName)
		}
		if !cmd.Flags().Changed(flagName) {
			continue
		}
		v, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return client.Request{}, err
		}
		switch p.In {
		case "path":
			req.Path = strings.ReplaceAll(req.Path, "{"+p.Name+"}", v)
		case "query":
			req.Query[p.Name] = v
		case "header":
			if req.Headers == nil {
				req.Headers = map[string]string{}
			}
			req.Headers[p.Name] = v
		}
	}
	return req, nil
}

func addParamFlag(cmd *cobra.Command, p metadata.Param) {
	name := paramFlagName(p.Name)
	help := p.Description
	if p.Required {
		help += " (required)"
	}
	cmd.Flags().String(name, "", help)
	if p.Required {
		_ = cmd.MarkFlagRequired(name)
	}
}

func paramFlagName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

func sortedResources(resources map[string]metadata.Resource) []metadata.Resource {
	var names []string
	for n := range resources {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]metadata.Resource, len(names))
	for i, n := range names {
		out[i] = resources[n]
	}
	return out
}

// serviceShortcuts collects the shortcuts declared on methods in a service.
func (r *Runtime) serviceShortcuts(svc metadata.Service) []shortcutRef {
	var out []shortcutRef
	for _, res := range svc.Resources {
		for _, m := range res.Methods {
			if m.Shortcut == "" {
				continue
			}
			out = append(out, shortcutRef{res: res, method: m, shortcut: m.Shortcut})
		}
	}
	return out
}

type shortcutRef struct {
	res      metadata.Resource
	method   metadata.Method
	shortcut string
}

func (r *Runtime) buildShortcutCmd(svc metadata.Service, sc shortcutRef) *cobra.Command {
	name := strings.TrimPrefix(sc.shortcut, "+")
	cmd := &cobra.Command{
		Use:   "+" + name,
		Short: sc.method.Summary,
		Long:  fmt.Sprintf("Shortcut for %s %s", sc.method.HTTPMethod, sc.method.Path),
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := r.buildRequest(cmd, sc.method)
			if err != nil {
				return err
			}
			return r.execute(cmd.Context(), cmd, req)
		},
	}
	for _, p := range sc.method.Params {
		addParamFlag(cmd, p)
	}
	return cmd
}
