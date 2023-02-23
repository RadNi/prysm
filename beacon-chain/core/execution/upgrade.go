package execution

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	state_native "github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/prysmaticlabs/prysm/v3/crypto/timelock"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// UpgradeToBellatrix updates inputs a generic state to return the version Bellatrix state.
// It inserts an empty `ExecutionPayloadHeader` into the state.
func UpgradeToBellatrix(state state.BeaconState) (state.BeaconState, error) {
	epoch := time.CurrentEpoch(state)

	currentSyncCommittee, err := state.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	nextSyncCommittee, err := state.NextSyncCommittee()
	if err != nil {
		return nil, err
	}
	prevEpochParticipation, err := state.PreviousEpochParticipation()
	if err != nil {
		return nil, err
	}
	currentEpochParticipation, err := state.CurrentEpochParticipation()
	if err != nil {
		return nil, err
	}
	inactivityScores, err := state.InactivityScores()
	if err != nil {
		return nil, err
	}

	primes1 := make([][]byte, 2)
	primes1[0] = make([]byte, 256)
	primes1[1] = make([]byte, 256)
	primes2 := make([][]byte, 2)
	primes2[0] = make([]byte, 256)
	primes2[1] = make([]byte, 256)
	s := &ethpb.BeaconStateBellatrix{
		GenesisTime:           state.GenesisTime(),
		GenesisValidatorsRoot: state.GenesisValidatorsRoot(),
		Slot:                  state.Slot(),
		Fork: &ethpb.Fork{
			PreviousVersion: state.Fork().CurrentVersion,
			CurrentVersion:  params.BeaconConfig().BellatrixForkVersion,
			Epoch:           epoch,
		},
		LatestBlockHeader:           state.LatestBlockHeader(),
		BlockRoots:                  state.BlockRoots(),
		StateRoots:                  state.StateRoots(),
		HistoricalRoots:             state.HistoricalRoots(),
		Eth1Data:                    state.Eth1Data(),
		Eth1DataVotes:               state.Eth1DataVotes(),
		Eth1DepositIndex:            state.Eth1DepositIndex(),
		Validators:                  state.Validators(),
		Balances:                    state.Balances(),
		RandaoMixes:                 state.RandaoMixes(),
		Slashings:                   state.Slashings(),
		PreviousEpochParticipation:  prevEpochParticipation,
		CurrentEpochParticipation:   currentEpochParticipation,
		JustificationBits:           state.JustificationBits(),
		PreviousJustifiedCheckpoint: state.PreviousJustifiedCheckpoint(),
		CurrentJustifiedCheckpoint:  state.CurrentJustifiedCheckpoint(),
		FinalizedCheckpoint:         state.FinalizedCheckpoint(),
		InactivityScores:            inactivityScores,
		CurrentSyncCommittee:        currentSyncCommittee,
		NextSyncCommittee:           nextSyncCommittee,
		LatestExecutionPayloadHeader: &enginev1.ExecutionPayloadHeader{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, 32),
			ReceiptsRoot:     make([]byte, 32),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			BlockNumber:      0,
			GasLimit:         0,
			GasUsed:          0,
			Timestamp:        0,
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, 32),
			TimelockPrivatekey: &enginev1.ElgamalPrivateKey{
				PublicKey: &enginev1.ElgamalPublicKey{
					P: make([]byte, 512),
					G: make([]byte, 512),
					Y: make([]byte, 512),
				},
				X: make([]byte, 512),
			},
			TimelockPublickey: &enginev1.ElgamalPublicKey{
				P: make([]byte, 512),
				G: make([]byte, 512),
				Y: make([]byte, 512),
			},
		},
		TimelockPrivatekey: &enginev1.ElgamalPrivateKey{
			PublicKey: &enginev1.ElgamalPublicKey{
				P: make([]byte, 512),
				G: make([]byte, 512),
				Y: make([]byte, 512),
			},
			X: make([]byte, 512),
		},
		TimelockPuzzle: ethpb.CopyTimelockPuzzle(timelock.PuzzlePlaceHolder()),
	}

	return state_native.InitializeFromProtoUnsafeBellatrix(s)
}
