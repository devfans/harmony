package quorum

import (
	"math/big"
	"fmt"

	"github.com/pkg/errors"

	"github.com/harmony-one/harmony/internal/params"
	"github.com/harmony-one/harmony/consensus/votepower"
	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
	"github.com/harmony-one/harmony/shard"
)

type NetworkID = shardingconfig.NetworkID
type ChainConfig = params.ChainConfig
type Schedule = shardingconfig.Schedule
type NetworkType = nodeconfig.NetworkType

// NewVerifier creates the quorum verifier for the given committee, epoch and whether the scenario
// is staking.
func NewVerifierWithConfig(
	networkType nodeconfig.NetworkType,
	schedule shardingconfig.Schedule,
	committee *shard.Committee, epoch *big.Int, isStaking bool) (Verifier, error) {
	if isStaking {
		return newStakeVerifierWithConfig(networkType, schedule ,committee, epoch)
	}
	return newUniformVerifier(committee)
}

// newStakeVerifier creates a stake verifier from the given committee
func newStakeVerifierWithConfig(networkType nodeconfig.NetworkType,
	schedule shardingconfig.Schedule,
	committee *shard.Committee, epoch *big.Int) (*stakeVerifier, error) {
	r, err := votepower.ComputeWithConfig(networkType, schedule, committee, epoch)
	if err != nil {
		return nil, errors.Wrap(err, "compute staking vote-power")
	}
	return &stakeVerifier{
		r: *r,
	}, nil
}

func GetNetworkConfigAndShardSchedule(id shardingconfig.NetworkID) (
	schedule shardingconfig.Schedule,
	networkType nodeconfig.NetworkType,
	config *params.ChainConfig,
	err error) {
	switch id {
	case shardingconfig.MainNet:
		schedule = shardingconfig.MainnetSchedule
		networkType = nodeconfig.Mainnet
		config = params.MainnetChainConfig
	case shardingconfig.TestNet:
		schedule = shardingconfig.TestnetSchedule
		networkType = nodeconfig.Testnet
		config = params.TestChainConfig
	default:
		err = fmt.Errorf("invalid network id: %v", id)
	}
	return
}