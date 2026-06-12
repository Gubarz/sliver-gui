package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetLoot returns all loot stored on the server.
func (a *App) GetLoot() (*clientpb.AllLoot, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.LootAll(context.Background(), &commonpb.Empty{})
}

// DownloadLoot fetches a loot item's content and saves it locally via a dialog.
func (a *App) DownloadLoot(lootID string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}
	loot, err := a.rpcClient.LootContent(context.Background(), &clientpb.Loot{ID: lootID})
	if err != nil {
		return "", err
	}
	if loot.File == nil {
		return "", fmt.Errorf("loot item has no file content")
	}
	fname := loot.File.Name
	if fname == "" {
		fname = loot.Name
	}
	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Loot",
		DefaultFilename: fname,
	})
	if err != nil {
		return "", fmt.Errorf("dialog error: %w", err)
	}
	if localPath == "" {
		return "", nil
	}
	if err := os.WriteFile(localPath, loot.File.Data, 0644); err != nil {
		return "", err
	}
	return localPath, nil
}

// GetCredentials returns the server's credential store.
func (a *App) GetCredentials() (*clientpb.Credentials, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.Creds(context.Background(), &commonpb.Empty{})
}

// RemoveLoot deletes a loot item by ID.
func (a *App) RemoveLoot(id string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.LootRm(context.Background(), &clientpb.Loot{ID: id})
	return err
}

// AddCredential stores a new credential.
func (a *App) AddCredential(username, plaintext, hash, collection string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.CredsAdd(context.Background(), &clientpb.Credentials{
		Credentials: []*clientpb.Credential{{
			Username:   username,
			Plaintext:  plaintext,
			Hash:       hash,
			Collection: collection,
		}},
	})
	return err
}

// RemoveCredential deletes a credential by ID.
func (a *App) RemoveCredential(id string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.CredsRm(context.Background(), &clientpb.Credentials{
		Credentials: []*clientpb.Credential{{ID: id}},
	})
	return err
}

// GetScreenshotData fetches a loot item's content and returns it as a base64 data URI.
func (a *App) GetScreenshotData(lootID string) (string, error) {
	if !a.connected {
		return "", fmt.Errorf("not connected")
	}
	loot, err := a.rpcClient.LootContent(context.Background(), &clientpb.Loot{ID: lootID})
	if err != nil {
		return "", err
	}
	if loot.File == nil {
		return "", fmt.Errorf("loot item has no file content")
	}

	encoded := base64.StdEncoding.EncodeToString(loot.File.Data)
	// Assume PNG for now, browsers usually sniff or don't care too much, but we can check extension
	ext := "png"
	if strings.HasSuffix(strings.ToLower(loot.Name), ".jpg") || strings.HasSuffix(strings.ToLower(loot.Name), ".jpeg") {
		ext = "jpeg"
	}

	return fmt.Sprintf("data:image/%s;base64,%s", ext, encoded), nil
}
