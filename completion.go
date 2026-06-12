package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	consts "github.com/bishopfox/sliver/client/constants"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/spf13/cobra"
)

// captureConsoleOutput redirects os.Stdout/os.Stderr to a pipe for the duration
// of fn and returns everything written. Sliver's CLI printf writes through the
// os.Stdout variable, so reassigning it captures command output.
func captureConsoleOutput(fn func() error) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	var buf bytes.Buffer
	copyDone := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buf, r)
		copyDone <- copyErr
	}()

	runErr := fn()

	os.Stdout, os.Stderr = origOut, origErr
	w.Close()
	copyErr := <-copyDone
	r.Close()

	if runErr == nil && copyErr != nil {
		runErr = copyErr
	}
	return buf.String(), runErr
}

// findCompletions is a helper that traverses a cobra command tree to find subcommands matching the given args.
func findCompletions(cmd *cobra.Command, args []string) []string {
	if len(args) == 0 {
		var comps []string
		for _, c := range cmd.Commands() {
			if !c.Hidden {
				comps = append(comps, c.Name())
			}
		}
		return comps
	}

	token := args[0]
	if len(args) == 1 {
		var comps []string
		for _, c := range cmd.Commands() {
			if !c.Hidden && strings.HasPrefix(c.Name(), token) {
				comps = append(comps, c.Name())
			}
		}
		return comps
	}

	// Navigate down the tree
	for _, c := range cmd.Commands() {
		if !c.Hidden && c.Name() == token {
			return findCompletions(c, args[1:])
		}
	}

	return nil
}

// CompleteCommand provides tab completion for nested commands (e.g., "armory search").
func (a *App) CompleteCommand(sessionID string, line string) ([]string, error) {
	if err := a.initConsole(); err != nil {
		return nil, err
	}

	a.cmdMu.Lock()
	defer a.cmdMu.Unlock()

	if sessionID != "" {
		a.sliverCon.App.SwitchMenu(consts.ImplantMenu)
	} else {
		a.sliverCon.App.SwitchMenu(consts.ServerMenu)
	}

	menu := a.sliverCon.App.ActiveMenu()
	if menu == nil || menu.Command == nil {
		return nil, nil
	}

	// Initialize commands dynamically if needed (the menu.Command tree is bound by SetCommands)
	// We need to ensure the tree has the commands available. In StartClient they were passed,
	// but reeflective handles them via an internal reset function. We can just add them manually
	// to the root command temporarily to introspect.
	var root *cobra.Command
	if sessionID != "" {
		root = &cobra.Command{}
		root.AddCommand(a.sliverCmds().Commands()...)
	} else {
		root = &cobra.Command{}
		root.AddCommand(a.serverCmds().Commands()...)
	}

	parts := strings.Split(line, " ")
	return findCompletions(root, parts), nil
}

func (a *App) CompletePath(sessionID string, partial string) ([]string, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Split into the directory text we keep, the prefix we match, and the
	// separator in use (Windows implants use backslashes).
	var dirText, prefix, sep string
	if i := strings.LastIndexAny(partial, "/\\"); i >= 0 {
		sep = string(partial[i])
		dirText = partial[:i+1]
		prefix = partial[i+1:]
	} else {
		sep = a.sessionSep(sessionID)
		prefix = partial
	}

	listPath := dirText
	if listPath == "" {
		listPath = "."
	}

	resp, err := a.rpcClient.Ls(context.Background(), &sliverpb.LsReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Path:    listPath,
	})
	if err != nil {
		return nil, err
	}

	lowerPrefix := strings.ToLower(prefix)
	var out []string
	for _, f := range resp.Files {
		if f.Name == "." || f.Name == ".." {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(f.Name), lowerPrefix) {
			continue
		}
		cand := dirText + f.Name
		if f.IsDir {
			cand += sep
		}
		out = append(out, cand)
	}
	return out, nil
}

// sessionSep returns the path separator the given session's OS uses.
func (a *App) sessionSep(sessionID string) string {
	sessions, err := a.rpcClient.GetSessions(context.Background(), &commonpb.Empty{})
	if err == nil {
		for _, s := range sessions.Sessions {
			if s.ID == sessionID && strings.EqualFold(s.OS, "windows") {
				return "\\"
			}
		}
	}
	return "/"
}
