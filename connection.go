package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bishopfox/sliver/client/assets"
	"github.com/bishopfox/sliver/client/transport"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
)

type PivotListenerSnapshot struct {
	ParentSessionID string
	ID              uint32
	Type            string
	BindAddress     string
	Pivots          []PivotConnectionSnapshot
}

type PivotConnectionSnapshot struct {
	PeerID        int64
	RemoteAddress string
}

// GetClientConfigs returns a list of available connection profile names.
func (a *App) GetClientConfigs() ([]string, error) {
	configs := assets.GetConfigs()
	var names []string
	for name := range configs {
		names = append(names, name)
	}
	return names, nil
}

// Connect connects to the teamserver using the specified profile.
// If profileName is empty, it uses the first available profile.
func (a *App) Connect(profileName string) error {
	configs := assets.GetConfigs()
	if len(configs) == 0 {
		return fmt.Errorf("no sliver configs found in ~/.sliver-client/configs")
	}

	var config *assets.ClientConfig
	if profileName != "" {
		cfg, ok := configs[profileName]
		if !ok {
			return fmt.Errorf("profile not found: %s", profileName)
		}
		config = cfg
	} else {
		// Just grab the first one
		for _, c := range configs {
			config = c
			break
		}
	}

	log.Printf("Connecting to %s:%d as %s", config.LHost, config.LPort, config.Operator)
	a.config = config

	rpcClient, grpcConn, err := transport.MTLSConnect(config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	a.rpcClient = rpcClient
	a.grpcConn = grpcConn
	a.connected = true

	// Begin streaming server events (session opened/closed, etc.) to the UI.
	a.startEventStream()

	return nil
}

// GetSessions returns a list of sessions
func (a *App) GetSessions() (*clientpb.Sessions, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	sessions, err := a.rpcClient.GetSessions(context.Background(), &commonpb.Empty{})
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetBeacons returns a list of beacons
func (a *App) GetBeacons() (*clientpb.Beacons, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	beacons, err := a.rpcClient.GetBeacons(context.Background(), &commonpb.Empty{})
	if err != nil {
		return nil, err
	}
	return beacons, nil
}

// GetPivots returns the global pivot graph
func (a *App) GetPivots() (*clientpb.PivotGraph, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.PivotGraph(context.Background(), &commonpb.Empty{})
}

// GetPivotListeners returns each session's actual pivot listener bind address.
// ActiveC2 contains the connecting implant's source port, which is not the
// listener port operators need to identify the pivot endpoint.
func (a *App) GetPivotListeners() ([]PivotListenerSnapshot, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	sessions, err := a.rpcClient.GetSessions(context.Background(), &commonpb.Empty{})
	if err != nil {
		return nil, err
	}

	snapshots := []PivotListenerSnapshot{}
	var snapshotsMu sync.Mutex
	var waitGroup sync.WaitGroup
	for _, session := range sessions.Sessions {
		sessionID := session.ID
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			listeners, err := a.rpcClient.PivotSessionListeners(
				ctx,
				&sliverpb.PivotListenersReq{
					Request: &commonpb.Request{SessionID: sessionID},
				},
			)
			if err != nil || listeners.GetResponse().GetErr() != "" {
				return
			}
			sessionSnapshots := []PivotListenerSnapshot{}
			for _, listener := range listeners.Listeners {
				snapshot := PivotListenerSnapshot{
					ParentSessionID: sessionID,
					ID:              listener.ID,
					Type:            listener.Type.String(),
					BindAddress:     listener.BindAddress,
					Pivots:          []PivotConnectionSnapshot{},
				}
				for _, pivot := range listener.Pivots {
					snapshot.Pivots = append(snapshot.Pivots, PivotConnectionSnapshot{
						PeerID:        pivot.PeerID,
						RemoteAddress: pivot.RemoteAddress,
					})
				}
				sessionSnapshots = append(sessionSnapshots, snapshot)
			}

			snapshotsMu.Lock()
			snapshots = append(snapshots, sessionSnapshots...)
			snapshotsMu.Unlock()
		}()
	}
	waitGroup.Wait()
	return snapshots, nil
}

// GetOperators returns the list of connected operators
func (a *App) GetOperators() (*clientpb.Operators, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.GetOperators(context.Background(), &commonpb.Empty{})
}
