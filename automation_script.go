package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/sobek"
)

const (
	defaultAutomationTimeout = 5 * time.Minute
	maxAutomationTimeout     = 1 * time.Hour
)

func (e *automationEngine) executeCommands(rule AutomationRule, target automationTarget) (string, []string, error) {
	ctx, cancel := automationContext(e.app.ctx, rule.TimeoutSeconds)
	defer cancel()

	var output strings.Builder
	var commands []string
	var runErr error
	for index, command := range rule.Commands {
		command = renderAutomationCommand(command, target)
		if strings.TrimSpace(command) == "" {
			continue
		}
		commands = append(commands, command)
		if index > 0 && rule.DelaySeconds > 0 {
			select {
			case <-ctx.Done():
				runErr = ctx.Err()
			case <-time.After(time.Duration(rule.DelaySeconds) * time.Second):
			}
			if runErr != nil {
				break
			}
		}
		result, err := e.runAutomationCommand(ctx, target, command)
		appendAutomationCommandOutput(&output, command, result, err)
		if err != nil {
			runErr = err
			if !rule.ContinueOnError {
				break
			}
		}
	}
	return output.String(), commands, runErr
}

func (e *automationEngine) executeJavaScript(
	rule AutomationRule,
	trigger string,
	target automationTarget,
) (string, []string, error) {
	ctx, cancel := automationContext(e.app.ctx, rule.TimeoutSeconds)
	defer cancel()

	vm := sobek.New()
	vm.SetMaxCallStackSize(2048)
	vm.SetFieldNameMapper(sobek.TagFieldNameMapper("json", true))

	var output strings.Builder
	var commands []string
	appendLog := func(values ...interface{}) {
		if output.Len() > 0 {
			output.WriteByte('\n')
		}
		for index, value := range values {
			if index > 0 {
				output.WriteByte(' ')
			}
			output.WriteString(fmt.Sprint(value))
		}
	}
	run := func(command string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		command = renderAutomationCommand(command, target)
		if strings.TrimSpace(command) == "" {
			return "", fmt.Errorf("command cannot be empty")
		}
		commands = append(commands, command)
		result, err := e.runAutomationCommand(ctx, target, command)
		appendAutomationCommandOutput(&output, command, result, err)
		return result, err
	}
	sleep := func(milliseconds int64) error {
		if milliseconds < 0 {
			return fmt.Errorf("sleep duration cannot be negative")
		}
		timer := time.NewTimer(time.Duration(milliseconds) * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}

	targetValue := map[string]string{
		"id": target.ID, "name": target.Name, "hostname": target.Hostname,
		"username": target.Username, "os": target.OS, "arch": target.Arch, "kind": target.Kind,
	}
	if err := vm.Set("target", targetValue); err != nil {
		return "", nil, err
	}
	if err := vm.Set("trigger", map[string]string{"type": trigger}); err != nil {
		return "", nil, err
	}
	if err := vm.Set("sliver", map[string]interface{}{
		"run": run, "log": appendLog, "sleep": sleep,
	}); err != nil {
		return "", nil, err
	}
	console := vm.NewObject()
	if err := console.Set("log", appendLog); err != nil {
		return "", nil, err
	}
	if err := vm.Set("console", console); err != nil {
		return "", nil, err
	}

	interruptDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt(ctx.Err())
		case <-interruptDone:
		}
	}()
	result, err := vm.RunString(rule.Script)
	close(interruptDone)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return output.String(), commands, fmt.Errorf(
				"JavaScript exceeded %s timeout",
				automationTimeout(rule.TimeoutSeconds),
			)
		}
		return output.String(), commands, fmt.Errorf("JavaScript: %w", err)
	}
	if result != nil && !sobek.IsUndefined(result) && !sobek.IsNull(result) {
		appendLog("Result:", result.Export())
	}
	return output.String(), commands, nil
}

func automationContext(parent context.Context, timeoutSeconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, automationTimeout(timeoutSeconds))
}

func automationTimeout(timeoutSeconds int) time.Duration {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultAutomationTimeout
	}
	if timeout > maxAutomationTimeout {
		timeout = maxAutomationTimeout
	}
	return timeout
}

func (e *automationEngine) runAutomationCommand(
	ctx context.Context,
	target automationTarget,
	command string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	result, taskID, err := e.app.runAutomationConsoleLine(target.ID, command)
	if err != nil || target.Kind != "beacon" {
		return result, err
	}

	result, _, err = e.app.awaitBeaconTask(ctx, target.ID, result, taskID)
	if err != nil {
		return result, err
	}
	return result, nil
}

func appendAutomationCommandOutput(output *strings.Builder, command, result string, err error) {
	if output.Len() > 0 {
		output.WriteString("\n\n")
	}
	fmt.Fprintf(output, "$ %s", command)
	if result != "" {
		output.WriteByte('\n')
		output.WriteString(result)
	}
	if err != nil {
		fmt.Fprintf(output, "\n[!] %v", err)
	}
}
