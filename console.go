package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bishopfox/sliver/client/command"
	"github.com/bishopfox/sliver/client/console"
	consts "github.com/bishopfox/sliver/client/constants"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

// RunSessionCommand runs a command line against a session using Sliver's REAL
// command set (the same cobra command tree the native client uses), captures
// its output, and returns it. There is no PTY and no readline: the console runs
// in CLI mode, so this is just "set the active session, run one command line,
// capture stdout." Per-session isolation is automatic because the active target
// is set per call and each implant holds its own state (cwd, etc.) server-side.
func (a *App) RunSessionCommand(sessionID string, line string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil
	}

	return a.runConsoleLine(sessionID, line)
}

func unsupportedConsoleCommand(root *cobra.Command, line string) string {
	args, err := shellquote.Split(line)
	if err != nil || len(args) == 0 {
		return ""
	}
	command, _, err := root.Find(args)
	if err != nil || command == nil {
		return ""
	}
	return nonInteractiveCommandReason(commandPath(command))
}

func commandPath(command *cobra.Command) string {
	var parts []string
	for current := command; current != nil && current.Parent() != nil; current = current.Parent() {
		parts = append(parts, current.Name())
	}
	for left, right := 0, len(parts)-1; left < right; left, right = left+1, right-1 {
		parts[left], parts[right] = parts[right], parts[left]
	}
	return strings.Join(parts, " ")
}

func nonInteractiveCommandReason(path string) string {
	switch path {
	case "shell":
		return "Open a dedicated Shell tab from the agent menu."
	case "shell attach":
		return "Attaching to a managed shell requires an interactive terminal."
	case "edit":
		return "Use the GUI file browser; edit requires an interactive terminal."
	case "hexedit":
		return "Hex editing requires an interactive terminal."
	case "docs":
		return "The documentation browser requires an interactive terminal."
	case "switch":
		return "Server switching requires an interactive terminal selector."
	case "ai":
		return "The AI conversation interface requires an interactive terminal."
	default:
		return ""
	}
}

// ListCommands returns the names of every available implant command, taken from
// Sliver's real command tree — this is what the console's tab-completion uses,
// so the list always matches what actually runs.
func (a *App) ListCommands() ([]string, error) {
	if err := a.initConsole(); err != nil {
		return nil, err
	}

	a.cmdMu.Lock()
	defer a.cmdMu.Unlock()

	root := a.sliverCmds()
	serverRoot := a.serverCmds()

	cmdMap := make(map[string]bool)
	var names []string

	for _, c := range root.Commands() {
		if c.Hidden {
			continue
		}
		cmdMap[c.Name()] = true
	}
	for _, c := range serverRoot.Commands() {
		if c.Hidden {
			continue
		}
		cmdMap[c.Name()] = true
	}

	for name := range cmdMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// initConsole lazily builds the Sliver client console in CLI mode (commands and
// RPC connection bound, but no interactive readline/PTY started).
func (a *App) initConsole() error {
	a.consoleOnce.Do(func() {
		con := console.NewConsole(false)
		serverCmds := command.ServerCommands(con, nil)
		sliverCmds := command.SliverCommands(con)

		details := &console.ConnectionDetails{Config: a.config}

		// run=false => IsCLI: binds commands + connection and returns without
		// starting readline. No global fd hijack, no PTY.
		if err := console.StartClient(con, a.rpcClient, a.grpcConn, details, serverCmds, sliverCmds, false, ""); err != nil {
			a.consoleErr = err
			return
		}
		a.sliverCon = con
		a.sliverCmds = sliverCmds
		a.serverCmds = serverCmds
	})
	return a.consoleErr
}

// runConsoleLine points the console at a session and runs one command line,
// capturing everything it prints. Serialized because it briefly redirects the
// process's stdout/stderr.
func (a *App) runConsoleLine(sessionID, line string) (string, error) {
	if err := a.initConsole(); err != nil {
		return "", err
	}

	a.cmdMu.Lock()
	defer a.cmdMu.Unlock()

	return a.runConsoleLineLocked(sessionID, line)
}

func (a *App) runAutomationConsoleLine(sessionID, line string) (string, string, error) {
	if err := a.initConsole(); err != nil {
		return "", "", err
	}

	a.cmdMu.Lock()
	defer a.cmdMu.Unlock()

	output, err := a.runConsoleLineLocked(sessionID, line)
	if err != nil {
		return output, "", err
	}

	matches := beaconTaskNoticePattern.FindStringSubmatch(output)
	if len(matches) != 2 {
		return output, "", nil
	}
	return output, a.removeBeaconTaskCallbackByPrefix(matches[1]), nil
}

func (a *App) runConsoleLineLocked(sessionID, line string) (string, error) {
	var sess *clientpb.Session
	var beacon *clientpb.Beacon
	var err error

	root := a.serverCmds()
	if sessionID != "" {
		root = a.sliverCmds()
	}
	if reason := unsupportedConsoleCommand(root, line); reason != "" {
		return "", fmt.Errorf("%s", reason)
	}

	if sessionID != "" {
		sess, beacon, err = a.findTarget(sessionID)
		if err != nil {
			return "", err
		}
	}

	// Point the console at this target (session OR beacon) and the implant menu.
	// Beacon commands are queued by upstream Sliver and return a task notice.
	if sessionID != "" {
		a.sliverCon.ActiveTarget.Set(sess, beacon)
		a.sliverCon.App.SwitchMenu(consts.ImplantMenu)
	} else {
		a.sliverCon.ActiveTarget.Set(nil, nil)
		a.sliverCon.App.SwitchMenu(consts.ServerMenu)
	}

	menu := a.sliverCon.App.ActiveMenu()
	out, runErr := captureConsoleOutput(func() error {
		return menu.RunCommandLine(context.Background(), line)
	})
	return strings.TrimRight(out, "\n"), runErr
}

// findTarget resolves an agent ID to either a session or a beacon.
func (a *App) findTarget(id string) (*clientpb.Session, *clientpb.Beacon, error) {
	ctx := context.Background()
	if sessions, err := a.rpcClient.GetSessions(ctx, &commonpb.Empty{}); err == nil {
		for _, s := range sessions.Sessions {
			if s.ID == id {
				return s, nil, nil
			}
		}
	}
	if beacons, err := a.rpcClient.GetBeacons(ctx, &commonpb.Empty{}); err == nil {
		for _, b := range beacons.Beacons {
			if b.ID == id {
				return nil, b, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("agent not found: %s", id)
}
