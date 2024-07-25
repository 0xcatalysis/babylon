package configurer

import (
	"fmt"
	"testing"

	"github.com/babylonchain/babylon/test/e2e/configurer/chain"
	"github.com/babylonchain/babylon/test/e2e/containers"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

type Configurer interface {
	ConfigureChains() error

	ClearResources() error

	GetChainConfig(chainIndex int) *chain.Config

	RunSetup() error

	RunValidators() error

	InstantiateBabylonContract() error

	RunHermesRelayerIBC() error

	// RunCosmosRelayerIBC configures IBC with Go relayer
	RunCosmosRelayerIBC() error

	RunIBCTransferChannel() error
}

var (
	// Last nodes are non validator nodes to serve as the ones using relayer. Out
	// validators are constantly sending bls transactions which make relayer operations
	// fail constantly

	// each started validator container corresponds to one of
	// the configurations below.
	validatorConfigsChainA = []*initialization.NodeConfig{
		{
			// this is a node that is used to state-sync from so its snapshot-interval
			// is frequent.
			Name:               "babylon-default-a-1",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   25,
			SnapshotKeepRecent: 10,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-a-2",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-a-3",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        false,
		},
	}
	validatorConfigsChainB = []*initialization.NodeConfig{
		{
			Name:               "babylon-default-b-1",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-b-2",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-b-3",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        false,
		},
	}
	ibcConfigChainA = &ibctesting.ChannelConfig{
		PortID:  zctypes.PortID,
		Order:   zctypes.Ordering,
		Version: zctypes.Version,
	}
	ibcConfigChainB = &ibctesting.ChannelConfig{
		PortID:  zctypes.PortID, // Will be replaced by the contract address in Phase 2 tests
		Order:   zctypes.Ordering,
		Version: zctypes.Version,
	}
)

// NewBTCTimestampingConfigurer returns a new Configurer for BTC timestamping service.
// TODO currently only one configuration is available. Consider testing upgrades
// when necessary
func NewBTCTimestampingConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	identifier := "btc-timestamp"
	containerManager, err := containers.NewManager(identifier, isDebugLogEnabled, false)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainA), ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainB), ibcConfigChainB),
		},
		withIBC(baseSetup), // base set up with IBC
		containerManager,
	), nil
}

func NewIBCTransferConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	identifier := "ibc-transfer"
	containerManager, err := containers.NewManager(identifier, isDebugLogEnabled, false)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainA), ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainB), ibcConfigChainB),
		},
		withIBCTransferChannel(baseSetup), // base set up with IBC
		containerManager,
	), nil
}

// NewBTCTimestampingPhase2Configurer returns a new Configurer for BTC timestamping service (phase 2).
func NewBTCTimestampingPhase2Configurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	identifier := "btc-timestamping"
	containerManager, err := containers.NewManager(identifier, isDebugLogEnabled, false)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainA), ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainB), ibcConfigChainB),
		},
		withPhase2IBC(baseSetup), // IBC setup (requires contract address)
		containerManager,
	), nil
}

// NewBTCTimestampingPhase2RlyConfigurer returns a new Configurer for BTC timestamping service (phase 2), using the Go relayer (rly).
func NewBTCTimestampingPhase2RlyConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	identifier := "btc-timestamping-rly"
	containerManager, err := containers.NewManager(identifier, isDebugLogEnabled, true)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainA), ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainB), ibcConfigChainB),
		},
		withPhase2RlyIBC(baseSetup), // IBC setup with wasmd and Go relayer
		containerManager,
	), nil
}

// NewBTCStakingConfigurer returns a new Configurer for BTC staking service
func NewBTCStakingConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	identifier := "btc-staking"
	containerManager, err := containers.NewManager(identifier, isDebugLogEnabled, false)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			// we only need 1 chain for testing BTC staking
			chain.New(t, containerManager, initialization.ChainAID, nodeConfigsWithIdentifier(identifier, validatorConfigsChainA), nil),
		},
		baseSetup, // base set up
		containerManager,
	), nil
}

func nodeConfigsWithIdentifier(identifier string, cfgs []*initialization.NodeConfig) []*initialization.NodeConfig {
	for _, cfg := range cfgs {
		cfg.Name = fmt.Sprintf("%s-%s", cfg.Name, identifier)
	}
	return cfgs
}
