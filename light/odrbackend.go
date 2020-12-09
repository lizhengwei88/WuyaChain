package light

import (
	"WuyaChain/common"
	"WuyaChain/common/errors"
	"fmt"
	rand2 "math/rand"
	"WuyaChain/core/store"
	"WuyaChain/log"
	"WuyaChain/p2p"
	"sync"
	"time"
)

var (
	errNoMorePeers   = errors.New("No peers found")
	errServiceQuited = errors.New("Service has quited")
)

type odrBackend struct {
	lock       sync.Mutex
	msgCh      chan *p2p.Message
	quitCh     chan struct{}
	requestMap map[uint32]chan odrResponse
	wg         sync.WaitGroup
	peers      *peerSet
	bcStore    store.BlockchainStore // used to validate the retrieved ODR object.
	log        *log.WuyaLog

	shard uint
}


func newOdrBackend(bcStore store.BlockchainStore, shard uint) *odrBackend {
	o := &odrBackend{
		msgCh:      make(chan *p2p.Message),
		requestMap: make(map[uint32]chan odrResponse),
		quitCh:     make(chan struct{}),
		bcStore:    bcStore,
		log:        log.GetLogger("odrBackend"),
		shard:      shard,
	}
	rand2.Seed(time.Now().UnixNano())
	return o
}


func (o *odrBackend) close() {
	select {
	case <-o.quitCh:
	default:
		close(o.quitCh)
	}

	o.wg.Wait()
	close(o.msgCh)
}


// retrieve retrieves the requested ODR object from remote peer with specified peer filter.
func (o *odrBackend) retrieveWithFilter(request odrRequest, filter peerFilter) (odrResponse, error) {
	reqID, ch, peerL, err := o.getReqInfo(filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		o.lock.Lock()
		delete(o.requestMap, reqID)
		close(ch)
		o.lock.Unlock()
	}()

	request.setRequestID(reqID)
	code, payload := request.code(), common.SerializePanic(request)
	for _, p := range peerL {
		o.log.Debug("peer send request, code = %s, payloadSizeBytes = %v", codeToStr(code), len(payload))
		if err = p2p.SendMessage(p.rw, code, payload); err != nil {
			o.log.Info("Failed to send message with peer %s", p.peerStrID)
			return nil, errors.NewStackedErrorf(err, "failed to send P2P message")
		}
	}

	timeout := time.NewTimer(msgWaitTimeout)
	defer timeout.Stop()

	select {
	case resp := <-ch:
		if err := resp.getError(); err != nil {
			return nil, errors.NewStackedError(err, "failed to handle ODR request on server side")
		}

		if err := resp.validate(request, o.bcStore); err != nil {
			return nil, errors.NewStackedError(err, "failed to valdiate ODR response")
		}

		return resp, nil
	case <-o.quitCh:
		return nil, errServiceQuited
	case <-timeout.C:

		return nil, fmt.Errorf("wait for msg reqid=%d timeout", reqID)
	}
}



func (o *odrBackend) getReqInfo(filter peerFilter) (uint32, chan odrResponse, []*peer, error) {
	peerL := o.peers.choosePeers(filter)
	if len(peerL) == 0 {
		return 0, nil, nil, errNoMorePeers
	}

	reqID := rand2.Uint32()
	ch := make(chan odrResponse)

	o.lock.Lock()
	if o.requestMap[reqID] != nil {
		panic("reqid conflicks")
	}

	o.requestMap[reqID] = ch
	o.lock.Unlock()
	return reqID, ch, peerL, nil
}