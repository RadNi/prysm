package timelock

import (
	"context"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/crypto/rsa"
	log "github.com/sirupsen/logrus"
	"time"
)

type Channels struct {
	TimelockSolutionFoundChannel chan *TimelockSolution
	TimelockRequestChannel       chan *TimelockRequest
}

type Service struct {
	ctx       context.Context
	channels  *Channels
	isRunning bool
	requests  []*TimelockRequest
	solutions map[types.Slot]*TimelockSolution
	stop      chan bool
}

func (s *Service) Channels() *Channels {
	return s.channels
}

func (s *Service) SetChannels(channels *Channels) {
	s.channels = channels
}

// NewService sets up a new instance with an ethclient when given a web3 endpoint as a string in the config.
func NewService(ctx context.Context, opts ...Option) (*Service, error) {
	return &Service{
		ctx:       ctx,
		isRunning: false,
		requests:  make([]*TimelockRequest, 0),
		solutions: make(map[types.Slot]*TimelockSolution),
		channels: &Channels{
			TimelockSolutionFoundChannel: make(chan *TimelockSolution),
			TimelockRequestChannel:       make(chan *TimelockRequest),
		},
		stop: make(chan bool),
	}, nil
}

// Start the powchain service's main event loop.
func (s *Service) Start() {

	// TODO checking slot number is not enough because of the reorg events. We need to store the solution based on
	// the puzzles parameters, not the slot number.

	for {
		select {
		case <-time.After(time.Second * 1):
			newRequests := make([]*TimelockRequest, 0)
			for _, r := range s.requests {
				r.Puzzle.T -= 1
				if r.Puzzle.T == 0 {
					log.Info("Solved a timelock puzzle for slot")
					sol := &TimelockSolution{
						Solution:   rsa.ToProtoRSAPrivatekey(rsa.ImportPrivateKey()),
						SlotNumber: r.SlotNumber,
					}
					s.solutions[r.SlotNumber] = sol
					if r.Res != nil {
						r.Res <- sol
					}
				} else {
					sol, prs := s.solutions[r.SlotNumber]
					if prs {
						if r.Res != nil {
							r.Res <- sol
						}
					} else {
						newRequests = append(newRequests, r)
					}
				}
			}
			s.requests = newRequests
		}
		select {
		case x := <-s.channels.TimelockRequestChannel:
			if x.SlotNumber <= 7 {
				if x.Res != nil {
					x.Res <- &TimelockSolution{
						Solution:   rsa.ToProtoRSAPrivatekey(rsa.ImportPrivateKey()),
						SlotNumber: x.SlotNumber,
					}
				}
			} else {
				sol, prs := s.solutions[x.SlotNumber]
				if prs {
					x.Res <- sol
				} else {
					s.requests = append(s.requests, x)
				}
			}
		case x := <-s.channels.TimelockSolutionFoundChannel:
			newRequests := make([]*TimelockRequest, 0)
			for _, r := range s.requests {
				if r.SlotNumber == x.SlotNumber {
					if r.Res != nil {
						r.Res <- x
					}
				} else {
					newRequests = append(newRequests, r)
				}
			}
			s.requests = newRequests
			s.solutions[x.SlotNumber] = x
		case <-s.stop:
			close(s.channels.TimelockRequestChannel)
			close(s.channels.TimelockSolutionFoundChannel)
			close(s.stop)
			return
		default:
			continue
		}
	}
}

// Stop the web3 service's main event loop and associated goroutines.
func (s *Service) Stop() error {
	s.stop <- true
	return nil
}

// Status is service health checks. Return nil or error.
func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	return nil
}
