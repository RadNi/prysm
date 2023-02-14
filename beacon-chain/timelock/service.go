package timelock

import (
	"context"
	"fmt"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/crypto/elgamal"
	"github.com/prysmaticlabs/prysm/v3/crypto/timelock"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
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
			s.cleanSolutions(sol.SlotNumber)
			r := s.solvers[sol.SlotNumber]
			if r.request.Res != nil {
				r.request.Res <- sol
			}
			delete(s.solvers, sol.SlotNumber)
		case req := <-s.channels.TimelockRequestChannel:
			if sol, prs := s.solutions[req.SlotNumber]; prs {
				fmt.Printf("sending already known answer\n")
				req.Res <- sol
			} else if req.SlotNumber <= 7 {
				if req.Res != nil {
					fmt.Printf("sending placeholder\n")
					req.Res <- &TimelockSolution{
						Solution:   elgamal.ImportPrivateKey(),
						SlotNumber: req.SlotNumber,
					}
				}
			} else if _, prs := s.solvers[req.SlotNumber]; !prs {
				fmt.Printf("First time getting this puzzle. Starting to solve it\n")
				solver := &timelockSolver{
					request: req,
					stop:    make(chan bool),
				}
				s.solvers[req.SlotNumber] = solver
				go solve(solver, puzzleSolved)
			} else if req.Puzzle == nil {
				fmt.Printf("Request was nil. Sending placeholder\n")
				sol := &TimelockSolution{
					Solution:   elgamal.PlaceHolderPrivateKey(),
					SlotNumber: req.SlotNumber,
				}
				s.solutions[req.SlotNumber] = sol
				if req.Res != nil {
					req.Res <- sol
				}
			} else {
				fmt.Printf("Timelock for slot %v is delayed. Sending placeholder\n", req.SlotNumber)
				// TODO handle the case in which the solution hasn't been found yet, but the solver is still working on it
				sol := &TimelockSolution{
					Solution:   elgamal.PlaceHolderPrivateKey(),
					SlotNumber: req.SlotNumber,
				}
				s.solutions[req.SlotNumber] = sol
				if req.Res != nil {
					req.Res <- sol
				}
			}
		case x := <-s.channels.TimelockSolutionFoundChannel:
			if x.Solution == nil {
				fmt.Printf("found a nill solution for slot %v. Replacing it with placeholder\n", x.SlotNumber)
				x.Solution = elgamal.PlaceHolderPrivateKey()
			}
			if rs, pres := s.solvers[x.SlotNumber]; pres {
				r := rs.request
				if r.Res != nil {
					r.Res <- x
				}
				rs.stop <- true
				delete(s.solvers, r.SlotNumber)
			}
			s.solutions[x.SlotNumber] = x
			s.cleanSolutions(x.SlotNumber)
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

func (s *Service) cleanSolutions(slot types.Slot) {
	updated := make(map[types.Slot]*TimelockSolution)
	for k, v := range s.solutions {
		if uint64(k) >= uint64(slot)-3 {
			updated[k] = v
		}
	}
	s.solutions = updated
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
	fmt.Printf("starting time lock for %v\n", solver.request.SlotNumber)

	p := solver.request.Puzzle
	fmt.Printf("attempting to solve %v\n", p.U)
	t := time.Now()
	sk := timelock.PuzzleSolve(p.U, p.V, p.N, p.G, p.T, p.H, int(solver.request.SlotNumber))
	fmt.Printf("duration for %v is %v\n", solver.request.SlotNumber, time.Now().Sub(t).Seconds())
	fmt.Printf("slot: %v ans: %v\n", solver.request.SlotNumber, sk)
	ch <- &TimelockSolution{
		Solution: &enginev1.ElgamalPrivateKey{
			PublicKey: &enginev1.ElgamalPublicKey{
				G: p.G,
				P: p.N,
				Y: p.U,
			},
			X: sk,
		},
		SlotNumber: solver.request.SlotNumber,
	}
	//T := new(big.Int).SetBytes(solver.request.Puzzle.T).Uint64()
	//select {
	//
	//case <-time.After(time.Second * time.Duration(T)):
	//	ch <- &TimelockSolution{
	//		Solution:   elgamal.ImportPrivateKey(),
	//		SlotNumber: solver.request.SlotNumber,
	//	}
	//	return
	//case <-solver.stop:
	//	return
	//}
}
