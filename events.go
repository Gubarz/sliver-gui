package main

import (
	"context"
	"log"

	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// startEventStream subscribes to the server event stream and forwards a
// simplified payload to the frontend as "sliver-event" runtime events.
func (a *App) startEventStream() {
	stream, err := a.rpcClient.Events(context.Background(), &commonpb.Empty{})
	if err != nil {
		log.Printf("failed to open event stream: %v", err)
		return
	}
	go func() {
		for {
			ev, err := stream.Recv()
			if err != nil {
				// Stream dropped (server gone / connection lost) — tell the UI
				// so it can fall back to its reconnect flow instead of relying
				// on a polling heartbeat.
				runtime.EventsEmit(a.ctx, "sliver-event", map[string]interface{}{"type": "stream-closed"})
				return
			}
			payload := map[string]interface{}{"type": ev.EventType}
			if ev.Session != nil {
				payload["sessionID"] = ev.Session.ID
				payload["hostname"] = ev.Session.Hostname
				payload["username"] = ev.Session.Username
			}
			if ev.Job != nil {
				payload["job"] = ev.Job.Name
			}
			if len(ev.Data) > 0 {
				payload["data"] = string(ev.Data)
			}
			if a.automation != nil {
				a.automation.handleSliverEvent(ev)
			}
			runtime.EventsEmit(a.ctx, "sliver-event", payload)
		}
	}()
}
