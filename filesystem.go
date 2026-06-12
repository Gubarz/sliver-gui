package main

import (
	"context"
	"fmt"

	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
)

// MakeDir creates a directory on the target session.
func (a *App) MakeDir(sessionID, path string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.Mkdir(context.Background(), &sliverpb.MkdirReq{
		Request: &commonpb.Request{SessionID: sessionID}, Path: path,
	})
	return err
}

// RemovePath deletes a file or directory on the target session.
func (a *App) RemovePath(sessionID, path string, recursive bool) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.Rm(context.Background(), &sliverpb.RmReq{
		Request: &commonpb.Request{SessionID: sessionID}, Path: path, Recursive: recursive, Force: true,
	})
	return err
}

// RenamePath moves/renames a file on the target session.
func (a *App) RenamePath(sessionID, src, dst string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.Mv(context.Background(), &sliverpb.MvReq{
		Request: &commonpb.Request{SessionID: sessionID}, Src: src, Dst: dst,
	})
	return err
}
