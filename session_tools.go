package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/bishopfox/sliver/util/encoders"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const defaultRPCTimeout = 60 * time.Second

// GetProcessList returns the running processes on the target session
// GetProcessList lists processes on a session. When fullInfo is true the implant
// performs deeper (noisier) enumeration to populate owner/architecture/session —
// the equivalent of `ps -f`, which is not opsec-safe and may trigger EDR.
func (a *App) GetProcessList(sessionID string, fullInfo bool) (*sliverpb.Ps, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	req := &sliverpb.PsReq{
		FullInfo: fullInfo,
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
	}

	return a.rpcClient.Ps(ctx, req)
}

// GetFileList returns the files in a specific directory on the target session
func (a *App) GetFileList(sessionID string, path string) (*sliverpb.Ls, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	if path == "" {
		path = "." // Default to current directory
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	req := &sliverpb.LsReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
		Path: path,
	}

	return a.rpcClient.Ls(ctx, req)
}

// KillProcess terminates a remote process
func (a *App) KillProcess(sessionID string, pid int32) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	req := &sliverpb.TerminateReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
		Pid:   pid,
		Force: true,
	}

	_, err := a.rpcClient.Terminate(ctx, req)
	return err
}

// TakeScreenshot captures the screen of the target and returns a base64 encoded string
func (a *App) TakeScreenshot(sessionID string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	req := &sliverpb.ScreenshotReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
	}

	resp, err := a.rpcClient.Screenshot(ctx, req)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(resp.Data), nil
}

// DownloadFile downloads a file from the target and saves it locally via a save dialog
func (a *App) DownloadFile(sessionID string, remotePath string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}

	// Ask user where to save it
	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save File",
	})
	if err != nil {
		return fmt.Errorf("dialog error: %w", err)
	}
	if localPath == "" {
		return nil // User cancelled
	}

	req := &sliverpb.DownloadReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
		Path: remotePath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	// This is a simplified unary download for small files
	resp, err := a.rpcClient.Download(ctx, req)
	if err != nil {
		return err
	}

	if resp.Encoder == "gzip" {
		decoded, err := new(encoders.Gzip).Decode(resp.Data)
		if err != nil {
			return fmt.Errorf("failed to decode gzip download: %w", err)
		}
		resp.Data = decoded
	}

	// We actually need to save the file using os.WriteFile
	return writeLocalFile(localPath, resp.Data)
}

func writeLocalFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// UploadFile uploads a local file to the target via an open dialog
func (a *App) UploadFile(sessionID string, remotePath string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}

	// Ask user which file to upload
	localPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File to Upload",
	})
	if err != nil {
		return fmt.Errorf("dialog error: %w", err)
	}
	if localPath == "" {
		return nil // User cancelled
	}

	return a.uploadLocalFile(sessionID, remotePath, localPath)
}

// UploadFiles uploads local paths supplied by the desktop file-drop runtime.
func (a *App) UploadFiles(sessionID string, remotePath string, localPaths []string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	for _, localPath := range localPaths {
		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(localPath), err)
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory; drop files only", filepath.Base(localPath))
		}
		if err := a.uploadLocalFile(sessionID, remotePath, localPath); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(localPath), err)
		}
	}
	return nil
}

func (a *App) uploadLocalFile(sessionID string, remotePath string, localPath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	encodedData, err := new(encoders.Gzip).Encode(data)
	if err != nil {
		return fmt.Errorf("failed to gzip compress upload: %w", err)
	}

	req := &sliverpb.UploadReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
			Timeout:   int64(defaultRPCTimeout / time.Second),
		},
		Path:     remotePath,
		FileName: filepath.Base(localPath),
		Data:     encodedData,
		Encoder:  "gzip",
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer cancel()

	resp, err := a.rpcClient.Upload(ctx, req)
	if err != nil {
		return err
	}
	if resp.Response != nil && resp.Response.Err != "" {
		return fmt.Errorf("%s", resp.Response.Err)
	}
	if resp.WrittenFiles == 0 {
		return fmt.Errorf("target did not report writing the uploaded file")
	}
	return nil
}
