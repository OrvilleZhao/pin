package pinlib

import (
	"errors"
	"io"
	"net"
	"sync"
)

// Client struct contains all fields for exchanging packets to the server through a TCP connection
type Client struct {
	// Unexported
	connections uint32        // connections holds info about how many connections to be made to remote pin
	iface       io.ReadWriter // handler for the tunneling interface

	// Exported
	Remote string       // Remote is the IP:PORT combination of the remote pin
	Hook   func() error // Hook is a function that runs immediately after the TCP connections are made
}

// NewClient is used to create a new client which makes 'connections' connections to the remote pin.
func NewClient(remote string, connections uint32, iface io.ReadWriter) (*Client, error) {
	// if number of connections is 0 it is pointless to run this VPN
	if connections == 0 {
		return nil, errors.New("connections should be greater than 0")
	}

	return &Client{iface: iface, Remote: remote, connections: connections, Hook: func() error { return nil }}, nil
}

// Start method makes TCP connections and starts the packet exchange from the local tunneling interface to the remote interface.
// This also makes Client struct to satisfy the pinlib.Peer interface.
func (c *Client) Start() error {
	// wait group to wait for all go routines to complete
	wg := &sync.WaitGroup{}
	ServerAddr, err := net.ResolveUDPAddr("udp", c.Remote)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		return err
	}

	conn.Write([]byte("DUMMY"))

	ex := &Exchanger{conn: conn, iface: c.iface}

	wg.Add(1)

	go ex.Start(wg)

	err = c.Hook()
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
