package light

import (
	"WuyaChain/common/errors"
	"WuyaChain/consensus"
	"WuyaChain/core/store"
	"WuyaChain/database"
	"WuyaChain/event"
	"WuyaChain/log"
)

func newLightChain(bcStore store.BlockchainStore, lightDB database.Database, odrBackend *odrBackend, engine consensus.Engine) (*LightChain, error) {
	chain := &LightChain{
		bcStore:    bcStore,
		odrBackend: odrBackend,
		engine:     engine,
		headerChangedEventManager: event.NewEventManager(),
		headRollbackEventManager: event.NewEventManager(),
		log: log.GetLogger("LightChain"),
	}

	currentHeaderHash, err := bcStore.GetHeadBlockHash()
	if err != nil {
		return nil, errors.NewStackedError(err, "failed to get HEAD block hash")
	}

	chain.currentHeader, err = bcStore.GetBlockHeader(currentHeaderHash)
	if err != nil {
		return nil, errors.NewStackedErrorf(err, "failed to get block header, hash = %v", currentHeaderHash)
	}

	td, err := bcStore.GetBlockTotalDifficulty(currentHeaderHash)
	if err != nil {
		return nil, errors.NewStackedErrorf(err, "failed to get block TD, hash = %v", currentHeaderHash)
	}

	chain.canonicalTD = td

	return chain, nil
}
