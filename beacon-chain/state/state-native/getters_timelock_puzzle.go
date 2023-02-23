package state_native

import (
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// LatestTimelockPuzzle corresponding to timelockPuzzle on the beacon chain.
func (b *BeaconState) LatestTimelockPuzzle() (*ethpb.TimelockPuzzle, error) {
	if b.version != version.Bellatrix {
		return nil, errNotSupported("LatestTimelockPuzzle", b.version)
	}

	//if b.currentEpochParticipation == nil {
	//	return nil, nil
	//}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.latestTimelockPuzzleVal(), nil
}

// latestTimelockPuzzleVal of the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) latestTimelockPuzzleVal() *ethpb.TimelockPuzzle {
	return ethpb.CopyTimelockPuzzle(b.latestTimelockPuzzle)
}
