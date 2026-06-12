package main

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/bishopfox/sliver/client/assets"
	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/rpcpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// App struct
type App struct {
	ctx       context.Context
	cancel    context.CancelFunc
	config    *assets.ClientConfig
	rpcClient rpcpb.SliverRPCClient
	grpcConn  *grpc.ClientConn
	connected bool

	// Lazily-built Sliver client console (CLI mode) used to run real commands.
	sliverCon   *console.SliverClient
	sliverCmds  func() *cobra.Command
	serverCmds  func() *cobra.Command
	consoleOnce sync.Once
	consoleErr  error
	cmdMu       sync.Mutex

	shellMu     sync.RWMutex
	shells      map[string]*guiShell
	nextShellID atomic.Uint64

	automation *automationEngine
}

// NewApp creates a new App application struct
func NewApp() *App {
	configureDefaultArmory()
	return &App{shells: make(map[string]*guiShell)}
}

func configureDefaultArmory() {
	const (
		publicKey = "RWSBpxpRWDrD7Fe+VvRE3c2VEDC2NK80rlNCj+BX0gz44Xw07r6KQD9L"
		repoURL   = "https://api.github.com/repos/sliverarmory/armory/releases"
	)

	if assets.DefaultArmoryPublicKey == "" {
		assets.DefaultArmoryPublicKey = publicKey
	}
	if assets.DefaultArmoryRepoURL == "" {
		assets.DefaultArmoryRepoURL = repoURL
	}
	assets.DefaultArmoryConfig.PublicKey = assets.DefaultArmoryPublicKey
	assets.DefaultArmoryConfig.RepoURL = assets.DefaultArmoryRepoURL
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	a.automation = newAutomationEngine(a)
	a.automation.start(a.ctx)
}

func (a *App) shutdown(context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
}
