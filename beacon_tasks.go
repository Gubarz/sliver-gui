package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	sliverTasks "github.com/bishopfox/sliver/client/command/tasks"
	"github.com/bishopfox/sliver/protobuf/clientpb"
)

const beaconTaskPollInterval = time.Second

var beaconTaskNoticePattern = regexp.MustCompile(`(?i)Tasked beacon .*\(([0-9a-f]{8})\)`)

// GetBeaconTasks returns the queue and history maintained by the upstream
// Sliver teamserver for a beacon.
func (a *App) GetBeaconTasks(beaconID string) (*clientpb.BeaconTasks, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	if strings.TrimSpace(beaconID) == "" {
		return nil, fmt.Errorf("beacon ID is required")
	}
	return a.rpcClient.GetBeaconTasks(
		context.Background(),
		&clientpb.Beacon{ID: beaconID},
	)
}

// GetBeaconTaskOutput retrieves a task and delegates response decoding and
// formatting to Sliver's native task renderer.
func (a *App) GetBeaconTaskOutput(taskID string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}
	if strings.TrimSpace(taskID) == "" {
		return "", fmt.Errorf("task ID is required")
	}
	if err := a.initConsole(); err != nil {
		return "", err
	}

	task, err := a.rpcClient.GetBeaconTaskContent(
		context.Background(),
		&clientpb.BeaconTask{ID: taskID},
	)
	if err != nil {
		return "", err
	}

	return a.renderBeaconTask(task)
}

// CancelBeaconTask cancels a task that has not yet been sent to its beacon.
func (a *App) CancelBeaconTask(taskID string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("task ID is required")
	}
	_, err := a.rpcClient.CancelBeaconTask(
		context.Background(),
		&clientpb.BeaconTask{ID: taskID},
	)
	return err
}

func (a *App) awaitBeaconTask(
	ctx context.Context,
	beaconID string,
	commandOutput string,
	taskID string,
) (string, bool, error) {
	matches := beaconTaskNoticePattern.FindStringSubmatch(commandOutput)
	if len(matches) != 2 {
		return commandOutput, false, nil
	}

	prefix := matches[1]
	var err error
	if taskID == "" {
		taskID, err = a.resolveBeaconTaskID(ctx, beaconID, prefix)
	}
	if err != nil {
		return commandOutput, true, err
	}
	a.removeBeaconTaskCallback(taskID)

	task, err := a.waitForBeaconTask(ctx, taskID)
	if err != nil {
		a.cancelPendingBeaconTask(taskID)
		return commandOutput, true, fmt.Errorf("beacon task %s: %w", shortTaskID(taskID), err)
	}

	state := strings.ToLower(task.State)
	if state != "completed" {
		if state == "" {
			state = "unknown"
		}
		return commandOutput, true, fmt.Errorf("beacon task %s %s", shortTaskID(task.ID), state)
	}

	rendered, err := a.renderBeaconTask(task)
	if err != nil {
		return commandOutput, true, err
	}
	return rendered, true, nil
}

func (a *App) resolveBeaconTaskID(ctx context.Context, beaconID, prefix string) (string, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		tasks, err := a.rpcClient.GetBeaconTasks(ctx, &clientpb.Beacon{ID: beaconID})
		if err != nil {
			return "", err
		}
		for _, task := range tasks.Tasks {
			if strings.HasPrefix(strings.ToLower(task.ID), strings.ToLower(prefix)) {
				return task.ID, nil
			}
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *App) waitForBeaconTask(ctx context.Context, taskID string) (*clientpb.BeaconTask, error) {
	ticker := time.NewTicker(beaconTaskPollInterval)
	defer ticker.Stop()

	for {
		task, err := a.rpcClient.GetBeaconTaskContent(ctx, &clientpb.BeaconTask{ID: taskID})
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(task.State) {
		case "completed", "failed", "canceled":
			return task, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *App) renderBeaconTask(task *clientpb.BeaconTask) (string, error) {
	if err := a.initConsole(); err != nil {
		return "", err
	}

	a.cmdMu.Lock()
	defer a.cmdMu.Unlock()

	output, err := captureConsoleOutput(func() error {
		sliverTasks.PrintTask(task, a.sliverCon)
		return nil
	})
	return strings.TrimRight(output, "\n"), err
}

func (a *App) removeBeaconTaskCallback(taskID string) {
	if a.sliverCon == nil {
		return
	}
	a.sliverCon.BeaconTaskCallbacksMutex.Lock()
	delete(a.sliverCon.BeaconTaskCallbacks, taskID)
	a.sliverCon.BeaconTaskCallbacksMutex.Unlock()
}

func (a *App) removeBeaconTaskCallbackByPrefix(prefix string) string {
	if a.sliverCon == nil {
		return ""
	}

	prefix = strings.ToLower(prefix)
	a.sliverCon.BeaconTaskCallbacksMutex.Lock()
	defer a.sliverCon.BeaconTaskCallbacksMutex.Unlock()
	for taskID := range a.sliverCon.BeaconTaskCallbacks {
		if strings.HasPrefix(strings.ToLower(taskID), prefix) {
			delete(a.sliverCon.BeaconTaskCallbacks, taskID)
			return taskID
		}
	}
	return ""
}

func (a *App) cancelPendingBeaconTask(taskID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task, err := a.rpcClient.GetBeaconTaskContent(ctx, &clientpb.BeaconTask{ID: taskID})
	if err != nil || strings.ToLower(task.State) != "pending" {
		return
	}
	_, _ = a.rpcClient.CancelBeaconTask(ctx, &clientpb.BeaconTask{ID: taskID})
}

func shortTaskID(taskID string) string {
	if index := strings.IndexByte(taskID, '-'); index >= 0 {
		return taskID[:index]
	}
	return taskID
}
