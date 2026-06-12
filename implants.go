package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func newImplantConfig(goos, goarch, format, c2url string, isBeacon bool, beaconInterval int64) (*clientpb.ImplantConfig, error) {
	formats := map[string]clientpb.OutputFormat{
		"exe":       clientpb.OutputFormat_EXECUTABLE,
		"shared":    clientpb.OutputFormat_SHARED_LIB,
		"shellcode": clientpb.OutputFormat_SHELLCODE,
		"service":   clientpb.OutputFormat_SERVICE,
	}
	outFmt, ok := formats[strings.ToLower(format)]
	if !ok {
		return nil, fmt.Errorf("unknown format %q (use exe, shared, shellcode, or service)", format)
	}

	c2url = strings.TrimSpace(c2url)
	if c2url == "" {
		return nil, fmt.Errorf("a C2 URL is required, e.g. mtls://10.0.0.1:443")
	}

	return &clientpb.ImplantConfig{
		GOOS:             strings.ToLower(goos),
		GOARCH:           strings.ToLower(goarch),
		Format:           outFmt,
		IsBeacon:         isBeacon,
		BeaconInterval:   beaconInterval,
		C2:               []*clientpb.ImplantC2{{Priority: 0, URL: c2url}},
		HTTPC2ConfigName: "default",
	}, nil
}

// GenerateImplant builds an implant and saves it locally via a save dialog.
// Returns the saved path. format is one of: exe, shared, shellcode, service.
func (a *App) GenerateImplant(goos, goarch, format, c2url, name string, isBeacon bool, beaconInterval int64) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}

	cfg, err := newImplantConfig(goos, goarch, format, c2url, isBeacon, beaconInterval)
	if err != nil {
		return "", err
	}

	resp, err := a.rpcClient.Generate(context.Background(), &clientpb.GenerateReq{
		Name:   name,
		Config: cfg,
	})
	if err != nil {
		return "", err
	}
	if resp.File == nil {
		return "", fmt.Errorf("server returned no implant file")
	}

	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Implant",
		DefaultFilename: resp.File.Name,
	})
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}
	if localPath == "" {
		return "", nil // cancelled
	}
	if err := os.WriteFile(localPath, resp.File.Data, 0755); err != nil {
		return "", fmt.Errorf("failed to save implant: %w", err)
	}
	return localPath, nil
}

// SaveProfile saves an implant profile to the teamserver.
func (a *App) SaveProfile(name, goos, goarch, format, c2url string, isBeacon bool, beaconInterval int64) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}

	cfg, err := newImplantConfig(goos, goarch, format, c2url, isBeacon, beaconInterval)
	if err != nil {
		return err
	}

	profile := &clientpb.ImplantProfile{
		Name:   name,
		Config: cfg,
	}

	_, err = a.rpcClient.SaveImplantProfile(context.Background(), profile)
	return err
}

// GenerateImplantFromProfile builds an implant using an existing profile configuration.
func (a *App) GenerateImplantFromProfile(profileConfigID string, name string, format int) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}

	cfg := &clientpb.ImplantConfig{
		ID:               profileConfigID, // This triggers the backend to reuse the profile's config
		HTTPC2ConfigName: "default",
		Format:           clientpb.OutputFormat(format),
	}

	resp, err := a.rpcClient.Generate(context.Background(), &clientpb.GenerateReq{
		Name:   name,
		Config: cfg,
	})
	if err != nil {
		return "", err
	}
	if resp.File == nil {
		return "", fmt.Errorf("server returned no implant file")
	}

	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Implant from Profile",
		DefaultFilename: resp.File.Name,
	})
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}
	if localPath == "" {
		return "", nil // cancelled
	}
	if err := os.WriteFile(localPath, resp.File.Data, 0755); err != nil {
		return "", fmt.Errorf("failed to save implant: %w", err)
	}
	return localPath, nil
}
