package app

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Tunnel forwards connections from a local port to a remote port (via a jump host)
type Tunnel struct {
	localURI  string
	jumpURI   string
	RemoteURI string
	Config    *ssh.ClientConfig
}

// NewTunnel creates a new Tunnel using the supplied config. If there are any problems with the
// config it will return an error.
func NewTunnel(c TunnelConfig) (*Tunnel, error) {

	// need three uris to tunnel
	if c.LocalURI == "" || c.JumpURI == "" || c.RemoteURI == "" {
		return nil, fmt.Errorf("need 3 uris to tunnel")
	}

	t := Tunnel{}
	t.localURI = c.LocalURI
	t.jumpURI = c.JumpURI
	t.RemoteURI = c.RemoteURI

	t.Config = &ssh.ClientConfig{}

	if c.Username != "" {
		t.Config.User = c.Username
	}

	t.Config.Auth = make([]ssh.AuthMethod, 0)

	// add password authentication if available
	if c.Password != "" {
		t.Config.Auth = append(t.Config.Auth, ssh.Password(c.Password))
	}

	// add identity key authentication if available
	if c.Identity != "" {
		userKey, err := readPrivateKey(c.Identity)

		if err != nil {
			return nil, fmt.Errorf("could not find specified identity file: %s", c.Identity)
		}

		t.Config.Auth = append(t.Config.Auth, userKey)
	}

	if len(t.Config.Auth) == 0 {
		return nil, fmt.Errorf("no authentication methods supplied")
	}

	// if insecure host checking is enabled do not check the known hosts file, otherwise
	// use a known hosts file to check the host key.
	if c.InsecureHostKeyChecking {
		t.Config.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }
	} else {
		if c.KnownHosts == "" {
			c.KnownHosts = path.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
		}

		hostKeyCallback, err := knownhosts.New(c.KnownHosts)

		if err != nil {
			return nil, fmt.Errorf("error reading known hosts file: %s ", err)
		}

		t.Config.HostKeyCallback = hostKeyCallback
	}

	return &t, nil
}

// return an ssh auth method using the specified key
func readPrivateKey(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)

	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}

// Start will start listening on the LocalURI for incoming connections. When it receives one,
// it will begin forwarding traffic between the local and remote connections.
func (tunnel *Tunnel) Start(ready chan bool) {
	listener, err := net.Listen("tcp", tunnel.localURI)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer listener.Close()

	// pass back a message so that we know it's OK to start sending connections
	ready <- true

	// handle any incoming connections on the local port
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}
		go tunnel.forward(conn)
	}
}

// forward traffic from local to remote and vice versa, routed through the given jump server.
// In order to stop the ssh lib from hanging when an error occurs, this function will log fatal
// if it encounters any problems. This is hard to test, so not ideal, but better than freezing
// if there's a bad authentication.
//
// see this ticket: https://github.com/golang/go/issues/21941
func (tunnel *Tunnel) forward(localConn net.Conn) {
	// make ssh connection to jump server
	serverConn, err := ssh.Dial("tcp", tunnel.jumpURI, tunnel.Config)
	if err != nil {
		log.Fatal("jump server dial error: ", err)
	}

	// make connection to target remote server from inside jump server
	remoteConn, err := serverConn.Dial("tcp", tunnel.RemoteURI)
	if err != nil {
		log.Fatal("remote server dial error: ", err)
	}

	// copies traffic from one network connection to another until an EOF is received
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s\n", err)
		}
	}

	// start two go processes - each one copying information from one connection to the other
	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}
