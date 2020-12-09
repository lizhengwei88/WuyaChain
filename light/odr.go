package light

import "WuyaChain/core/store"

const (
	blockRequestCode uint16 = 10 + iota
	blockResponseCode
	addTxRequestCode
	addTxResponseCode
	trieRequestCode
	trieResponseCode
	receiptRequestCode
	receiptResponseCode
	txByHashRequestCode
	txByHashResponseCode
	debtRequestCode
	debtResponseCode
	protocolMsgCodeLength // protocolMsgCodeLength always defined in the end.
)


type odrRequest interface {
	getRequestID() uint32                                         // get the random request ID.
	setRequestID(requestID uint32)                                // set the random request ID.
	code() uint16                                                 // get request code.
	handle(lp *LightProtocol) (respCode uint16, resp odrResponse) // handle the request and return response to remote peer.
}

type odrResponse interface {
	getRequestID() uint32                                             // get the random request ID.
	setRequestID(requestID uint32)                                    // set the random request ID.
	getError() error                                                  // get the response error if any.
	setError(err error)                                               // set the response error.
	validate(request odrRequest, bcStore store.BlockchainStore) error // validate the retrieved response.
}

// OdrItem is base struct for ODR request and response.
type OdrItem struct {
	ReqID uint32 // random request ID that generated dynamically
	Error string // response error
}