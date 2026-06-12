package main

import (
	"context"
	"fmt"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
)

// GetCertificates returns the server's mTLS certificates
func (a *App) GetCertificates() (*clientpb.CertificateInfo, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.GetCertificateInfo(context.Background(), &clientpb.CertificatesReq{})
}

// GetWebsites returns the server's active websites
func (a *App) GetWebsites() (*clientpb.Websites, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	return a.rpcClient.Websites(context.Background(), &commonpb.Empty{})
}

// GetAliases returns the loaded aliases
func (a *App) GetAliases() (interface{}, error) {
	// Not easily available via RPC since it's client side
	// Return empty for now to avoid compilation issues, we'll parse command output instead
	return nil, nil
}
