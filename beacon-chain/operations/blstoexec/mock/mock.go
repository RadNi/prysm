package mock

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	eth "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Changes []*eth.SignedBLSToExecutionChange
}

// PendingBLSToExecChanges --
func (m *PoolMock) PendingBLSToExecChanges() ([]*eth.SignedBLSToExecutionChange, error) {
	return m.Changes, nil
}

// BLSToExecChangesForInclusion --
func (m *PoolMock) BLSToExecChangesForInclusion(_ state.BeaconState) ([]*eth.SignedBLSToExecutionChange, error) {
	return m.Changes, nil
}

// InsertBLSToExecChange --
func (m *PoolMock) InsertBLSToExecChange(change *eth.SignedBLSToExecutionChange) {
	m.Changes = append(m.Changes, change)
}

// MarkIncluded --
func (*PoolMock) MarkIncluded(_ *eth.SignedBLSToExecutionChange) error {
	panic("implement me")
}

// ValidatorExists --
func (*PoolMock) ValidatorExists(_ types.ValidatorIndex) bool {
	panic("implement me")
}
