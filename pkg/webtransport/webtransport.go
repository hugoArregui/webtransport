package webtransport

import (
	"io"
	"sync"

	"github.com/pion/sctp"
)

// Ref: https://wicg.github.io/web-transport/

type ReadableStream io.Reader
type WritableStream io.Writer
type Datagram []byte

type WebTransportState int

const (
	WebTransportStateNew = iota + 1
	WebTransportStateConnecting
	WebTransportStateConnected
	WebTransportStateClosed
	WebTransportStateFailed
)

type WebTransportCloseInfo struct {
	ErrorCode uint16
	Reason    string
}

// TODO: struct or interface?: I think we need both,
type WebTransport interface {
	Close(closeInfo *WebTransportCloseInfo) error
	State() WebTransportState
	//   attribute EventHandler onstatechange;
	//   attribute EventHandler onerror;
}

type WebTransportStream interface {
	StreamID() uint64
	Transport() WebTransport
}

type StreamAbortInfo struct {
	ErrorCode uint16
}

type OutgoingStream interface {
	Writable() WritableStream
	//   readonly attribute Promise<StreamAbortInfo> writingAborted;
	AbortWriting(abortInfo *StreamAbortInfo)
}

type SendStream interface {
	WebTransportStream
	OutgoingStream
}

type IncomingStream interface {
	Readable() ReadableStream
	// readonly attribute Promise<StreamAbortInfo> readingAborted;
	AbortReading(abortInfo *StreamAbortInfo)
}

type ReceiveStream interface {
	WebTransportStream
	IncomingStream
}

type BidirectionalStream interface {
	WebTransportStream
	OutgoingStream
	IncomingStream
}

type SendStreamParameters struct {
	DisableRetransmissions bool
}

type UnidirectionalStreamsTransport interface {
	CreateSendStream(parameters SendStreamParameters) (SendStream, error)
	ReceiveStreams() []IncomingStream
}

type BidirectionalStreamsTransport interface {
	CreateBidirectionalStream() (BidirectionalStream, error)
	ReceiveBidirectionalStreams() []BidirectionalStream
}

type DatagramTransport interface {
	MaxDatagramSize() uint16
	SendDatagrams() WritableStream
	ReceiveDatagrams() ReadableStream
}

type SCTPStream struct {
	stream    *sctp.Stream
	transport WebTransport
}

func (s *SCTPStream) StreamID() uint64 {
	return uint64(s.stream.StreamIdentifier())
}

func (s *SCTPStream) Transport() WebTransport {
	return s.transport
}

func (s *SCTPStream) Writable() WritableStream {
	return s.stream
}

func (s *SCTPStream) AbortWriting(abortInfo *StreamAbortInfo) {
	//TODO
}

func (s *SCTPStream) Readable() ReadableStream {
	return s.stream
}

func (s *SCTPStream) AbortReading(abortInfo *StreamAbortInfo) {
	//TODO
}

// NOTE: SCTPTransport implements UnidirectionalStreamsTransport,
// BidirectionalStreamsTransport, DatagramTransport and WebTransport
type SCTPTransport struct {
	mux                          sync.Mutex
	nextStreamID                 uint16
	outgoingStreams              []OutgoingStream
	receivedStreams              []IncomingStream
	receivedBidirectionalStreams []BidirectionalStream
	// sentDatagrams
	association *sctp.Association
}

func NewSCTPTransport(association *sctp.Association) *SCTPTransport {
	return &SCTPTransport{
		association: association,
	}
}

func (t *SCTPTransport) CreateSendStream(parameters SendStreamParameters) (SendStream, error) {
	// TODO check transport state
	t.mux.Lock()
	defer t.mux.Unlock()

	//TODO: reliability parameters
	stream, err := t.createSCTPStream()
	if err != nil {
		return nil, err
	}

	// https://wicg.github.io/web-transport/#add-sendstream
	t.outgoingStreams = append(t.outgoingStreams, stream)
	return stream, nil
}

func (t *SCTPTransport) CreateBidirectionalStream() (BidirectionalStream, error) {
	// TODO check transport state
	t.mux.Lock()
	defer t.mux.Unlock()

	//TODO: parameters
	stream, err := t.createSCTPStream()
	if err != nil {
		return nil, err
	}

	// https://wicg.github.io/web-transport/#add-the-bidirectionalstream
	t.outgoingStreams = append(t.outgoingStreams, stream)
	t.receivedBidirectionalStreams = append(t.receivedBidirectionalStreams, stream)
	return stream, nil
}

func (t *SCTPTransport) SendDatagrams() WritableStream {
	//TODO: I don't totally get this, should we always have an open stream for this kind of thing?

	// sent out of order, unreliably, and have a limited maximum size. Datagrams
	// are encrypted and congestion controlled.
	// If too many datagrams are queued
	// because the stream is not being read quickly enough, drop datagrams to avoid
	// queueing. Implementations should drop older datagrams in favor of newer
	// datagrams. The number of datagrams to queue should be kept small enough to
	// avoid adding significant latency to packet delivery when the stream is being
	// read slowly (due to the reader being slow) but large enough to avoid
	// dropping packets when for the stream is not read for short periods of time
	// (due to the reader being paused).
	return nil
}

func (t *SCTPTransport) ReceiveStreams() []IncomingStream {
	// TODO: but.. where are the received streams accepted?
	// should be accept streams here or have goroutine or what?
	return t.receivedStreams
}

func (t *SCTPTransport) ReceiveBidirectionalStreams() []BidirectionalStream {
	return t.receivedBidirectionalStreams
}

func (t *SCTPTransport) ReceiveDatagrams() ReadableStream {
	// TODO
	return nil
}

func (t *SCTPTransport) MaxDatagramSize() uint16 {
	// TODO: ideally the association should return it's MTU
	return 1000
}

func (t *SCTPTransport) Close(closeInfo *WebTransportCloseInfo) error {
	//TODO parameters
	return t.association.Close()
}

//NOTE: caller should hold the lock
func (t *SCTPTransport) createSCTPStream() (*SCTPStream, error) {
	// TODO check transport state
	streamID := t.nextStreamID
	t.nextStreamID++

	sctpStream, err := t.association.OpenStream(streamID, sctp.PayloadTypeWebRTCBinary)
	if err != nil {
		return nil, err
	}

	stream := &SCTPStream{transport: t, stream: sctpStream}
	return stream, nil
}

func (t *SCTPTransport) State() WebTransportState {
	// TODO: should we store state or "translate" the association state?
	return WebTransportState(0)
}
