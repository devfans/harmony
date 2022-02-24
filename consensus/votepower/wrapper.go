package votepower

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/harmony-one/harmony/shard"
	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
	"github.com/harmony-one/harmony/numeric"
)

// Compute creates a new roster based off the shard.SlotList with target schedule and network type
func ComputeWithConfig(
	networkType nodeconfig.NetworkType,
	schedule shardingconfig.Schedule, subComm *shard.Committee, epoch *big.Int,
	) (*Roster, error) {
	if epoch == nil {
		return nil, errors.New("nil epoch for roster compute")
	}
	roster, staked := NewRoster(subComm.ShardID), subComm.Slots

	for i := range staked {
		if e := staked[i].EffectiveStake; e != nil {
			roster.TotalEffectiveStake = roster.TotalEffectiveStake.Add(*e)
		} else {
			roster.HMYSlotCount++
		}
	}

	asDecHMYSlotCount := numeric.NewDec(roster.HMYSlotCount)
	// TODO Check for duplicate BLS Keys
	ourPercentage := numeric.ZeroDec()
	theirPercentage := numeric.ZeroDec()
	var lastStakedVoter *AccommodateHarmonyVote

	harmonyPercent := schedule.InstanceForEpoch(epoch).HarmonyVotePercent()
	externalPercent := schedule.InstanceForEpoch(epoch).ExternalVotePercent()

	// Testnet incident recovery
	// Make harmony nodes having 70% voting power for epoch 73314
	if networkType == nodeconfig.Testnet && epoch.Cmp(big.NewInt(73305)) >= 0 &&
		epoch.Cmp(big.NewInt(73490)) <= 0 {
		harmonyPercent = numeric.MustNewDecFromStr("0.70")
		externalPercent = numeric.MustNewDecFromStr("0.40") // Make sure consensus is always good.
	}

	for i := range staked {
		member := AccommodateHarmonyVote{
			PureStakedVote: PureStakedVote{
				EarningAccount: staked[i].EcdsaAddress,
				Identity:       staked[i].BLSPublicKey,
				GroupPercent:   numeric.ZeroDec(),
				EffectiveStake: numeric.ZeroDec(),
				RawStake:       numeric.ZeroDec(),
			},
			OverallPercent: numeric.ZeroDec(),
			IsHarmonyNode:  false,
		}

		// Real Staker
		if e := staked[i].EffectiveStake; e != nil {
			member.EffectiveStake = member.EffectiveStake.Add(*e)
			member.GroupPercent = e.Quo(roster.TotalEffectiveStake)
			member.OverallPercent = member.GroupPercent.Mul(externalPercent)
			theirPercentage = theirPercentage.Add(member.OverallPercent)
			lastStakedVoter = &member
		} else { // Our node
			member.IsHarmonyNode = true
			member.OverallPercent = harmonyPercent.Quo(asDecHMYSlotCount)
			member.GroupPercent = member.OverallPercent.Quo(harmonyPercent)
			ourPercentage = ourPercentage.Add(member.OverallPercent)
		}

		if _, ok := roster.Voters[staked[i].BLSPublicKey]; !ok {
			roster.Voters[staked[i].BLSPublicKey] = &member
		} else {
			fmt.Printf("Duplicate BLS key found, blsKey: %s\n", staked[i].BLSPublicKey.Hex())
		}
	}

	if !(networkType == nodeconfig.Testnet && epoch.Cmp(big.NewInt(73305)) >= 0 &&
		epoch.Cmp(big.NewInt(73490)) <= 0) {

		// NOTE Enforce voting power sums to one,
		// give diff (expect tiny amt) to last staked voter
		if diff := numeric.OneDec().Sub(
			ourPercentage.Add(theirPercentage),
		); !diff.IsZero() && lastStakedVoter != nil {
			lastStakedVoter.OverallPercent =
				lastStakedVoter.OverallPercent.Add(diff)
			theirPercentage = theirPercentage.Add(diff)
		}

		if lastStakedVoter != nil &&
			!ourPercentage.Add(theirPercentage).Equal(numeric.OneDec()) {
			return nil, ErrVotingPowerNotEqualOne
		}
	}

	roster.OurVotingPowerTotalPercentage = ourPercentage
	roster.TheirVotingPowerTotalPercentage = theirPercentage
	for _, slot := range subComm.Slots {
		roster.OrderedSlots = append(roster.OrderedSlots, slot.BLSPublicKey)
	}
	return roster, nil
}

