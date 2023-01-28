package timelock

import (
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

type TimelockRequest struct {
	SlotNumber types.Slot
	Puzzle     *ethpb.TimelockPuzzle
	Res        chan *TimelockSolution
}
type TimelockNewPuzzle struct {
	SlotNumber types.Slot
	Puzzle     *ethpb.TimelockPuzzle
}
type TimelockSolution struct {
	SlotNumber types.Slot
	Solution   *enginev1.RSAPrivateKey
}
