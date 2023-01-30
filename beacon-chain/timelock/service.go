package timelock

import (
	"context"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/crypto/rsa"
	"time"
)

type Channels struct {
	TimelockSolutionFoundChannel chan *TimelockSolution
	TimelockRequestChannel       chan *TimelockRequest
}

type timelockSolver struct {
	request *TimelockRequest
	stop    chan bool
}

type Service struct {
	ctx       context.Context
	channels  *Channels
	isRunning bool
	solvers   map[types.Slot]*timelockSolver
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
		solvers:   make(map[types.Slot]*timelockSolver),
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

	puzzleSolved := make(chan *TimelockSolution)

	for {
		select {
		case sol := <-puzzleSolved:
			s.solutions[sol.SlotNumber] = sol
			r := s.solvers[sol.SlotNumber]
			if r.request.Res != nil {
				r.request.Res <- sol
			}
			delete(s.solvers, sol.SlotNumber)
		case req := <-s.channels.TimelockRequestChannel:
			if req.SlotNumber <= 7 {
				if req.Res != nil {
					req.Res <- &TimelockSolution{
						Solution:   rsa.ToProtoRSAPrivatekey(rsa.ImportPrivateKey()),
						SlotNumber: req.SlotNumber,
					}
				}
			} else {

				if sol, prs := s.solutions[req.SlotNumber]; prs {
					req.Res <- sol
				} else {
					if _, prs := s.solvers[req.SlotNumber]; !prs {
						solver := &timelockSolver{
							request: req,
							stop:    make(chan bool),
						}
						s.solvers[req.SlotNumber] = solver
						go solve(solver, puzzleSolved)
					}
				}
			}
		case x := <-s.channels.TimelockSolutionFoundChannel:
			if rs, pres := s.solvers[x.SlotNumber]; pres {
				r := rs.request
				if r.Res != nil {
					r.Res <- x
				}
				rs.stop <- true
				delete(s.solvers, r.SlotNumber)
			}
			s.solutions[x.SlotNumber] = x
		case <-s.stop:
			close(s.channels.TimelockRequestChannel)
			close(s.channels.TimelockSolutionFoundChannel)
			for _, solver := range s.solvers {
				solver.stop <- true
				close(solver.stop)
			}
			close(puzzleSolved)
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

func solve(solver *timelockSolver, ch chan *TimelockSolution) {
	select {

	case <-time.After(time.Second * time.Duration(solver.request.Puzzle.T)):
		ch <- &TimelockSolution{
			Solution:   rsa.ToProtoRSAPrivatekey(rsa.ImportPrivateKey()),
			SlotNumber: solver.request.SlotNumber,
		}
		return
	case <-solver.stop:
		return
	}
}
