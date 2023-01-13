package state_native

import (
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	_ "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// SetTimelockPrivatekey for the beacon state.
func (b *BeaconState) SetTimelockPrivatekey(val *enginev1.RSAPrivateKey) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version != version.Bellatrix {
		return errNotSupported("SetTimelockPrivatekey", b.version)
	}
	b.latestTimelockPrivateKey = val
	return nil
}
