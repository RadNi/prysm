package state_native

import (
	_ "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	eth "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// SetTimelockPuzzle for the beacon state.
func (b *BeaconState) SetTimelockPuzzle(val *eth.TimelockPuzzle) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version != version.Bellatrix {
		return errNotSupported("SetTimelockPrivatekey", b.version)
	}
	b.latestTimelockPuzzle = val
	return nil
}
