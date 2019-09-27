package webtransport_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/pion/logging"
	"github.com/pion/sctp"
	"github.com/pion/webtransport/pkg/webtransport"
	"github.com/stretchr/testify/require"
)

func createServer() (*sctp.Association, error) {
	laddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10002")
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10001")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", laddr, raddr)
	if err != nil {
		return nil, err
	}

	return sctp.Server(sctp.Config{
		NetConn:       conn,
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	})
}

func createClient() (*sctp.Association, error) {
	laddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10001")
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10002")
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP("udp4", laddr, raddr)
	if err != nil {
		return nil, err
	}

	return sctp.Client(sctp.Config{
		NetConn:       c,
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	})
}

// from: https://wicg.github.io/web-transport/#example-unreliable-delivery
func TestUnreliableDeliver(t *testing.T) {
	go func() {
		association, err := createServer()
		require.NoError(t, err)

		transport := webtransport.NewSCTPTransport(association)

		stream, err := transport.
			CreateSendStream(webtransport.SendStreamParameters{DisableRetransmissions: true})
		require.NoError(t, err)

		msg := make([]byte, 10)
		fmt.Println("WRITING")
		_, err = stream.Writable().Write(msg)
		require.NoError(t, err)
	}()

	go func() {
		association, err := createClient()
		require.NoError(t, err)

		transport := webtransport.NewSCTPTransport(association)

		for _, stream := range transport.ReceiveStreams() {
			buf := make([]byte, 10)
			fmt.Println("READING")
			_, err = stream.Readable().Read(buf)
			require.NoError(t, err)
		}
	}()

	fmt.Println("BLOCK FOREVER")
	select {}
}
