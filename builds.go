package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetImplantBuilds returns previously generated implant builds.
func (a *App) GetImplantBuilds() (*clientpb.ImplantBuilds, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.ImplantBuilds(context.Background(), &commonpb.Empty{})
}

// DeleteImplantBuild removes a stored implant build by name.
func (a *App) DeleteImplantBuild(name string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.DeleteImplantBuild(context.Background(), &clientpb.DeleteReq{Name: name})
	return err
}

// GetProfiles returns saved implant profiles.
func (a *App) GetProfiles() (*clientpb.ImplantProfiles, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.ImplantProfiles(context.Background(), &commonpb.Empty{})
}

// DeleteProfile removes a saved implant profile by name.
func (a *App) DeleteProfile(name string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.DeleteImplantProfile(context.Background(), &clientpb.DeleteReq{Name: name})
	return err
}

// RegenerateImplant re-downloads a previously built implant by name and saves
// it locally via a dialog.
func (a *App) RegenerateImplant(name string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}
	resp, err := a.rpcClient.Regenerate(context.Background(), &clientpb.RegenerateReq{ImplantName: name})
	if err != nil {
		return "", err
	}
	if resp.File == nil {
		return "", fmt.Errorf("no build artifact found for %q", name)
	}
	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Implant",
		DefaultFilename: resp.File.Name,
	})
	if err != nil || localPath == "" {
		return "", err
	}
	if err := os.WriteFile(localPath, resp.File.Data, 0755); err != nil {
		return "", err
	}
	return localPath, nil
}
