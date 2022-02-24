package consensus

import (
	"fmt"
	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
	"github.com/harmony-one/harmony/internal/params"
)

type NetworkID = shardingconfig.NetworkID
type ChainConfig = params.ChainConfig
type Schedule = shardingconfig.Schedule
type NetworkType = nodeconfig.NetworkType

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
