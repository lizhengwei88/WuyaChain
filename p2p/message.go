package p2p

import "time"

// Message exposed for high level layer to receive
type Message struct {
	Code       uint16 // message code, defined in each protocol
	Payload    []byte
	ReceivedAt time.Time
}

// MsgReader interface
type MsgReader interface {
	// ReadMsg read a message. It will block until send the message out or get errors
	ReadMsg() (*Message, error)
}

// MsgWriter interface
type MsgWriter interface {
	// WriteMsg sends a message. It will block until the message's
	// Payload has been consumed by the other end.
	//
	// Note that messages can be sent only once because their
	// payload reader is drained.
	WriteMsg(*Message) error
}

// MsgReadWriter provides reading and writing of encoded messages.
// Implementations should ensure that ReadMsg and WriteMsg can be
// called simultaneously from multiple goroutines.
type MsgReadWriter interface {
	MsgReader
	MsgWriter
}

// SendMessage send message to peer
func SendMessage(writer MsgWriter, code uint16, payload []byte) error {
	msg := Message{
		Code:    code,
		Payload: payload,
	}

	return writer.WriteMsg(&msg)
}