package pinlib

import (
	"fmt"
	"io"
	"net"
)

// Server struct contains all fields for exchanging packets to the client through a TCP connection
type Server struct {
	server  *net.UDPConn
	iface   io.ReadWriter
	clients map[string]*UDPTunnel
}

// NewServer method is used to create a new server struct with a given listening address
func NewServer(addr string, iface io.ReadWriter) (*Server, error) {
	ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error while resolving : %s", err)
	}

	ln, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		return nil, fmt.Errorf("Error while listening : %s", err)
	}

	return &Server{server: ln, iface: iface, clients: make(map[string]*UDPTunnel, 0)}, nil
}

// Start method accepts TCP connections from a client and starts the packet exchange from the local tunneling interface to the remote client
// This also makes Server struct to satisfy the pinlib.Peer interface.
func (s *Server) Start() error {

	p := make([]byte, 1500)

	for {
		n, addr, err := s.server.ReadFromUDP(p)
		if err != nil {
			continue
		}

		cid, ok := s.clients[addr.String()]
		if !ok {
			// New clients
			fmt.Println("Received handshake from :", addr)
			pr, pw := io.Pipe()
			s.clients[addr.String()] = &UDPTunnel{pipein: pw, pipeout: pr, remoteAddr: addr, conn: s.server, close: false}
			cid = s.clients[addr.String()]
			ex := &Exchanger{conn: cid, iface: s.iface}
			go ex.Start(nil)
		}
		cid.pipein.Write(p[:n])
	}

	return nil
}

type UDPTunnel struct {
	pipein     io.Writer
	pipeout    io.Reader
	remoteAddr *net.UDPAddr
	conn       *net.UDPConn
	close      bool
}

func (u *UDPTunnel) Write(p []byte) (int, error) {
	return u.conn.WriteToUDP(p, u.remoteAddr)
}

func (u *UDPTunnel) Read(p []byte) (int, error) {
	return u.pipeout.Read(p)
}
