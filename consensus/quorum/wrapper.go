package quorum

import (
	"math/big"

	"github.com/pkg/errors"

	"github.com/harmony-one/harmony/consensus/votepower"
	"github.com/harmony-one/harmony/shard"
	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
)

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
