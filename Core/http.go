package wireproxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// requestHeaderTimeout bounds how long a client may take to send the request
// line and headers, so idle or slowloris connections cannot pile up.
const requestHeaderTimeout = 30 * time.Second

const proxyAuthHeaderKey = "Proxy-Authorization"

type HTTPServer struct {
	config *HTTPConfig

	auth CredentialValidator
	dial func(network, address string) (net.Conn, error)

	authRequired bool
	tlsRequired  bool
}

func (s *HTTPServer) authenticate(req *http.Request) (int, error) {
	if !s.authRequired {
		return 0, nil
	}

	auth := req.Header.Get(proxyAuthHeaderKey)
	if auth == "" {
		return http.StatusProxyAuthRequired, fmt.Errorf("%s", http.StatusText(http.StatusProxyAuthRequired))
	}

	enc := strings.TrimPrefix(auth, "Basic ")
	str, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return http.StatusNotAcceptable, fmt.Errorf("decode username and password failed: %w", err)
	}
	pairs := bytes.SplitN(str, []byte(":"), 2)
	if len(pairs) != 2 {
		return http.StatusLengthRequired, fmt.Errorf("username and password format invalid")
	}
	if s.auth.Valid(string(pairs[0]), string(pairs[1])) {
		return 0, nil
	}
	return http.StatusUnauthorized, fmt.Errorf("username and password not matching")
}

func (s *HTTPServer) handleConn(req *http.Request, conn net.Conn) (peer net.Conn, err error) {
	addr := req.Host
	if !strings.Contains(addr, ":") {
		port := "443"
		addr = net.JoinHostPort(addr, port)
	}

	peer, err = s.dial("tcp", addr)
	if err != nil {
		return peer, fmt.Errorf("tun tcp dial failed: %w", err)
	}

	_, err = conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		_ = peer.Close()
		peer = nil
	}

	return
}

func (s *HTTPServer) handle(req *http.Request) (peer net.Conn, err error) {
	addr := req.Host
	if !strings.Contains(addr, ":") {
		port := "80"
		addr = net.JoinHostPort(addr, port)
	}

	peer, err = s.dial("tcp", addr)
	if err != nil {
		return peer, fmt.Errorf("tun tcp dial failed: %w", err)
	}

	err = req.Write(peer)
	if err != nil {
		_ = peer.Close()
		peer = nil
		return peer, fmt.Errorf("conn write failed: %w", err)
	}

	return
}

func (s *HTTPServer) serve(conn net.Conn) {
	handedOff := false
	defer func() {
		if !handedOff {
			_ = conn.Close()
		}
	}()

	_ = conn.SetReadDeadline(time.Now().Add(requestHeaderTimeout))
	var rd = bufio.NewReader(conn)
	req, err := http.ReadRequest(rd)
	if err != nil {
		log.Printf("read request failed: %s\n", err)
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	code, err := s.authenticate(req)
	if err != nil {
		resp := responseWith(req, code)
		if code == http.StatusProxyAuthRequired {
			resp.Header.Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
		}
		_ = resp.Write(conn)
		log.Println(err)
		return
	}

	var peer net.Conn
	switch req.Method {
	case http.MethodConnect:
		peer, err = s.handleConn(req, conn)
	case http.MethodGet:
		peer, err = s.handle(req)
	default:
		_ = responseWith(req, http.StatusMethodNotAllowed).Write(conn)
		log.Printf("unsupported protocol: %s\n", req.Method)
		return
	}
	if err != nil {
		log.Printf("dial proxy failed: %s\n", err)
		return
	}
	if peer == nil {
		log.Println("dial proxy failed: peer nil")
		return
	}

	handedOff = true
	go func() {
		defer func() { _ = conn.Close() }()
		defer func() { _ = peer.Close() }()

		_, _ = io.Copy(conn, peer)
	}()

	go func() {
		defer func() { _ = conn.Close() }()
		defer func() { _ = peer.Close() }()

		_, _ = io.Copy(peer, conn)
	}()
}

func (s *HTTPServer) listen(network, addr string) (net.Listener, error) {
	if s.tlsRequired {
		cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return nil, err
		}

		return tls.Listen(network, addr, &tls.Config{Certificates: []tls.Certificate{cert}})
	}

	return net.Listen(network, addr)
}

// Serve runs the accept loop on an existing listener until the listener is closed.
func (s *HTTPServer) Serve(server net.Listener) {
	defer func(server net.Listener) {
		_ = server.Close()
	}(server)
	for {
		conn, err := server.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			// Transient errors (e.g. EMFILE) must not kill the listener.
			log.Printf("accept request failed: %s\n", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		go func(conn net.Conn) {
			s.serve(conn)
		}(conn)
	}
}
