package core

import (
	"pkg/mod/github.com/ethereum/go-ethereum@v1.9.12/common/prque"
	"sync"
)

// New creates an Istanbul consensus core
func New(backend istanbul.Backend, config *istanbul.Config) Engine {
	c := &core{
		config:             config,
		address:            backend.Address(),
		state:              StateAcceptRequest,
		handlerWg:          new(sync.WaitGroup),
		logger:             log.GetLogger("ibft_core"),
		backend:            backend,
		backlogs:           make(map[common.Address]*prque.Prque),
		backlogsMu:         new(sync.Mutex),
		pendingRequests:    prque.New(),
		pendingRequestsMu:  new(sync.Mutex),
		consensusTimestamp: time.Time{},
		roundMeter:         metrics.GetOrRegisterMeter("consensus/istanbul/core/round", nil),
		sequenceMeter:      metrics.GetOrRegisterMeter("consensus/istanbul/core/sequence", nil),
		consensusTimer:     metrics.GetOrRegisterTimer("consensus/istanbul/core/consensus", nil),
	}
	c.validateFn = c.checkValidatorSignature
	return c
}