package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
)

// GetJobs returns the server's active jobs (listeners, etc.).
func (a *App) GetJobs() (*clientpb.Jobs, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.GetJobs(context.Background(), &commonpb.Empty{})
}

// KillJob stops a running job (e.g. a listener) by ID.
func (a *App) KillJob(id uint32) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	_, err := a.rpcClient.KillJob(context.Background(), &clientpb.KillJobReq{ID: id})
	return err
}

// StartListener starts a C2 listener of the given protocol. `domains` is a
// comma-separated list used only by the DNS listener.
func (a *App) StartListener(protocol, host string, port uint32, domains string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	ctx := context.Background()

	switch strings.ToLower(protocol) {
	case "mtls":
		_, err := a.rpcClient.StartMTLSListener(ctx, &clientpb.MTLSListenerReq{Host: host, Port: port})
		return err
	case "http":
		_, err := a.rpcClient.StartHTTPListener(ctx, &clientpb.HTTPListenerReq{Host: host, Port: port, Secure: false})
		return err
	case "https":
		_, err := a.rpcClient.StartHTTPSListener(ctx, &clientpb.HTTPListenerReq{Host: host, Port: port, Secure: true})
		return err
	case "dns":
		var doms []string
		for _, d := range strings.Split(domains, ",") {
			if d = strings.TrimSpace(d); d != "" {
				doms = append(doms, d)
			}
		}
		_, err := a.rpcClient.StartDNSListener(ctx, &clientpb.DNSListenerReq{Domains: doms, Host: host, Port: port})
		return err
	default:
		return fmt.Errorf("unknown listener protocol: %s", protocol)
	}
}
