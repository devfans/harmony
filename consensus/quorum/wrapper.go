package quorum

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/harmony-one/harmony/consensus/votepower"
	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
	"github.com/harmony-one/harmony/internal/params"
	bls_core "github.com/harmony-one/bls/ffi/go/bls"
	"github.com/harmony-one/harmony/crypto/bls"
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

// Verify if all desired harmony accounts is set in slots of new epoch
func CheckHarmonyAccountsInSlots(instance shardingconfig.Instance, shardID int, slots shard.SlotList) (err error) {
	shardNum := int(instance.NumShards())
	hmyAccounts := instance.HmyAccounts()
	for j := 0; j < instance.NumHarmonyOperatedNodesPerShard(); j++ {
		index := shardID + j*shardNum // The initial account to use for genesis nodes
		pub := &bls_core.PublicKey{}
		pub.DeserializeHexStr(hmyAccounts[index].BLSPublicKey)
		pubKey := bls.SerializedPublicKey{}
		pubKey.FromLibBLSPublicKey(pub)
		found := false
		for _, slot := range slots {
			if slot.BLSPublicKey == pubKey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unexpected harmony node not found in slots, pubkey: %s", pubKey.Hex())
		}
	}
	return
}