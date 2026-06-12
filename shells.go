package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bishopfox/sliver/client/core"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	maxShellOutputBytes = 4 * 1024 * 1024
	shellTrimTarget     = 3 * 1024 * 1024
	shellEventBatchSize = 512 * 1024
	shellEventInterval  = 100 * time.Millisecond
)

type ShellInfo struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionID"`
	Path      string `json:"path"`
	PID       uint32 `json:"pid"`
	PTY       bool   `json:"pty"`
}

type guiShell struct {
	info   ShellInfo
	tunnel *core.TunnelIO
	mu     sync.RWMutex
	output []byte
}

func (a *App) StartShell(sessionID, path string, enablePTY bool, rows, cols uint32) (*ShellInfo, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	if err := a.initConsole(); err != nil {
		return nil, err
	}

	session, beacon, err := a.findTarget(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || beacon != nil {
		return nil, fmt.Errorf("interactive shells require a live session")
	}

	osName := strings.ToLower(session.OS)
	enablePTY = enablePTY && (osName == "linux" || osName == "darwin")
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 80
	}

	rpcTunnel, err := a.rpcClient.CreateTunnel(context.Background(), &sliverpb.Tunnel{
		SessionID: sessionID,
	})
	if err != nil {
		return nil, err
	}
	tunnel := core.GetTunnels().Start(rpcTunnel.TunnelID, rpcTunnel.SessionID)

	response, err := a.rpcClient.Shell(context.Background(), &sliverpb.ShellReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(59 * time.Second),
		},
		Path:      strings.TrimSpace(path),
		EnablePTY: enablePTY,
		Rows:      rows,
		Cols:      cols,
		TunnelID:  tunnel.ID,
	})
	if err != nil {
		if closeErr := a.closeShellTunnel(tunnel.ID, sessionID); closeErr != nil {
			log.Printf("shell: failed to close tunnel after shell error: %v", closeErr)
		}
		return nil, err
	}
	if response.Response != nil && response.Response.Err != "" {
		if closeErr := a.closeShellTunnel(tunnel.ID, sessionID); closeErr != nil {
			log.Printf("shell: failed to close tunnel after remote shell error: %v", closeErr)
		}
		return nil, fmt.Errorf("%s", response.Response.Err)
	}

	id := fmt.Sprintf("shell-%d", a.nextShellID.Add(1))
	shellPath := strings.TrimSpace(path)
	if shellPath == "" {
		shellPath = response.Path
	}
	info := ShellInfo{
		ID:        id,
		SessionID: sessionID,
		Path:      shellPath,
		PID:       response.Pid,
		PTY:       enablePTY,
	}

	a.shellMu.Lock()
	a.shells[id] = &guiShell{info: info, tunnel: tunnel}
	a.shellMu.Unlock()

	go a.readShell(id, tunnel)
	return &info, nil
}

func (a *App) WriteShell(id, data string) error {
	shell, err := a.getShell(id)
	if err != nil {
		return err
	}
	if !shell.info.PTY {
		data = strings.ReplaceAll(data, "\x03", "")
		if data == "" {
			return nil
		}
	}
	_, err = shell.tunnel.Write([]byte(data))
	return err
}

// InterruptShell interrupts an interactive shell. PTY-backed shells understand
// ETX. Windows shells are stdin/stdout pipes, where ETX becomes literal command
// text, so the only reliable interruption is to terminate the shell process.
func (a *App) InterruptShell(id string) (bool, error) {
	shell, err := a.getShell(id)
	if err != nil {
		return false, err
	}
	if shell.info.PTY {
		_, err = shell.tunnel.Write([]byte{3})
		return false, err
	}

	a.shellMu.Lock()
	delete(a.shells, id)
	a.shellMu.Unlock()
	return true, a.closeShellTunnel(shell.tunnel.ID, shell.info.SessionID)
}

func (a *App) GetShellOutput(id string) (string, error) {
	a.shellMu.RLock()
	shell := a.shells[id]
	a.shellMu.RUnlock()
	if shell == nil {
		return "", fmt.Errorf("shell %q was not found", id)
	}
	shell.mu.RLock()
	defer shell.mu.RUnlock()
	return string(shell.output), nil
}

func (a *App) ResizeShell(id string, rows, cols uint32) error {
	shell, err := a.getShell(id)
	if err != nil {
		return err
	}
	if !shell.info.PTY || rows == 0 || cols == 0 {
		return nil
	}
	_, err = a.rpcClient.ShellResize(context.Background(), &sliverpb.ShellResizeReq{
		Request:  &commonpb.Request{SessionID: shell.info.SessionID, Timeout: int64(9 * time.Second)},
		Rows:     rows,
		Cols:     cols,
		TunnelID: shell.tunnel.ID,
	})
	return err
}

func (a *App) CloseShell(id string) error {
	a.shellMu.Lock()
	shell := a.shells[id]
	delete(a.shells, id)
	a.shellMu.Unlock()
	if shell == nil {
		return nil
	}

	_, _ = shell.tunnel.Write([]byte("exit\n"))
	return a.closeShellTunnel(shell.tunnel.ID, shell.info.SessionID)
}

func (a *App) getShell(id string) (*guiShell, error) {
	a.shellMu.RLock()
	shell := a.shells[id]
	a.shellMu.RUnlock()
	if shell == nil || !shell.tunnel.IsOpen() {
		return nil, fmt.Errorf("shell %q is closed", id)
	}
	return shell, nil
}

func (a *App) readShell(id string, tunnel *core.TunnelIO) {
	buffer := make([]byte, 64*1024)
	outputEvents := make(chan []byte, 64)
	eventsDone := make(chan struct{})
	go a.emitShellOutput(id, outputEvents, eventsDone)

	var readErr error
	for {
		n, err := tunnel.Read(buffer)
		if n > 0 {
			chunk := append([]byte(nil), buffer[:n]...)
			a.shellMu.RLock()
			shell := a.shells[id]
			a.shellMu.RUnlock()
			if shell != nil {
				shell.mu.Lock()
				shell.output = appendBoundedShellOutput(shell.output, chunk)
				shell.mu.Unlock()
			}
			outputEvents <- chunk
		}
		if err != nil {
			if err != io.EOF {
				readErr = err
			}
			break
		}
	}

	close(outputEvents)
	<-eventsDone
	if readErr != nil {
		runtime.EventsEmit(a.ctx, "shell-output", map[string]interface{}{
			"id":    id,
			"error": readErr.Error(),
		})
	}
	runtime.EventsEmit(a.ctx, "shell-output", map[string]interface{}{"id": id, "closed": true})
}

func (a *App) emitShellOutput(id string, chunks <-chan []byte, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(shellEventInterval)
	defer ticker.Stop()

	pending := make([]byte, 0, shellEventBatchSize)
	flush := func() {
		if len(pending) == 0 {
			return
		}
		runtime.EventsEmit(a.ctx, "shell-output", map[string]interface{}{
			"id":   id,
			"data": string(pending),
		})
		pending = pending[:0]
	}

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				flush()
				return
			}
			pending = append(pending, chunk...)
			if len(pending) >= shellEventBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func appendBoundedShellOutput(output, chunk []byte) []byte {
	output = append(output, chunk...)
	if len(output) <= maxShellOutputBytes {
		return output
	}

	start := len(output) - shellTrimTarget
	if newline := bytes.IndexByte(output[start:], '\n'); newline >= 0 {
		start += newline + 1
	}
	trimmed := make([]byte, len(output)-start)
	copy(trimmed, output[start:])
	return trimmed
}

func (a *App) closeShellTunnel(tunnelID uint64, sessionID string) error {
	core.GetTunnels().Close(tunnelID)
	_, err := a.rpcClient.CloseTunnel(context.Background(), &sliverpb.Tunnel{
		TunnelID:  tunnelID,
		SessionID: sessionID,
	})
	return err
}
