package main

import (
	"context"
	"fmt"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
)

// KillAgent kills a session or beacon by ID.
func (a *App) KillAgent(id string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	sess, beacon, err := a.findTarget(id)
	if err != nil {
		return err
	}
	req := &commonpb.Request{}
	if sess != nil {
		req.SessionID = sess.ID
	} else {
		req.BeaconID = beacon.ID
	}
	_, err = a.rpcClient.Kill(context.Background(), &sliverpb.KillReq{Request: req, Force: true})
	return err
}

// RemoveBeacon deletes a beacon record and all of its tasks from the
// teamserver. Unlike KillAgent, this does not require the implant to check in.
func (a *App) RemoveBeacon(id string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	if id == "" {
		return fmt.Errorf("beacon ID is required")
	}
	_, err := a.rpcClient.RmBeacon(context.Background(), &clientpb.Beacon{ID: id})
	return err
}

// RenameAgent sets a friendly name on a session or beacon.
func (a *App) RenameAgent(id, name string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	sess, beacon, err := a.findTarget(id)
	if err != nil {
		return err
	}
	req := &clientpb.RenameReq{Name: name}
	if sess != nil {
		req.SessionID = sess.ID
	} else {
		req.BeaconID = beacon.ID
	}
	_, err = a.rpcClient.Rename(context.Background(), req)
	return err
}

// GetVersion returns the teamserver version info.
func (a *App) GetVersion() (*clientpb.Version, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.GetVersion(context.Background(), &commonpb.Empty{})
}
