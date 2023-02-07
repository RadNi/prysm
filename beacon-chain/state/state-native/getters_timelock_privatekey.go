package state_native

import (
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// LatestTimelockPrivatekey corresponding to timelockPrivatekey on the beacon chain.
func (b *BeaconState) LatestTimelockPrivatekey() (*enginev1.ElgamalPrivateKey, error) {
	if b.version != version.Bellatrix {
		return nil, errNotSupported("LatestTimelockPrivatekey", b.version)
	}

	//if b.currentEpochParticipation == nil {
	//	return nil, nil
	//}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.latestTimelockPrivatekeyVal(), nil
}

// latestTimelockPrivatekeyVal of the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) latestTimelockPrivatekeyVal() *enginev1.ElgamalPrivateKey {
	return ethpb.CopyElgamalPrivatekey(b.latestTimelockPrivateKey)
}
