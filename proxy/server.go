package proxy

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"

	"github.com/google/uuid"

	buf2 "github.com/eugene-ivanov-hash/mitm-proxy/buf"
	"github.com/eugene-ivanov-hash/mitm-proxy/rule"
)

const (
	BufSize = 1024 * 32
)

type Server struct {
	caCert        *x509.Certificate
	caKey         any
	requestRules  []*rule.Rule
	responseRules []*rule.Rule
}

func NewProxySslServer(rootCa, rootKey string, requestRules []*rule.Rule, responseRules []*rule.Rule) *Server {
	caCert, caKey, err := loadX509KeyPair(rootCa, rootKey)
	if err != nil {
		log.Fatal("Error loading CA certificate/key:", err)
	}
	log.Printf("loaded CA certificate and key; IsCA=%v\n", caCert.IsCA)
	return &Server{
		caCert:        caCert,
		caKey:         caKey,
		requestRules:  requestRules,
		responseRules: responseRules,
	}
}

func (p Server) HandleTLS(conn net.Conn) {
	br := bufio.NewReader(conn)

	bc := buf2.NewBufferedConn(conn, br)
	defer bc.Close()

	peek, err := br.Peek(7)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to peek HTTP/1.1 5: %v", err))
		return
	}

	var (
		host string
		r    *http.Request
	)
	if peek[0] == 'C' && string(peek) == "CONNECT" {
		r, err = http.ReadRequest(br)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to read HTTP/1.1 5: %v", err))
			return
		}
		host = r.Host
		if _, err = bc.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
			slog.Error(fmt.Sprintf("Failed to write HTTP/1.1 200 OK: %v", err))
			return
		}

		peek, err = br.Peek(7)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to peek HTTP/1.1 5: %v", err))
			return
		}
	}

	if peek[0] == 0x16 {
		p.handleHTTPS(bc, host)
	} else {
		p.handleHTTP(bc)
	}
}

func (p Server) handleHTTP(clientConn net.Conn) {
	p.handle(clientConn, false)
}

func (p Server) handleHTTPS(clientConn net.Conn, host string) {
	tlsCert := getTlsCert(host, p.caCert, p.caKey)
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:       tls.VersionTLS13,
		Certificates:     []tls.Certificate{*tlsCert},
	}

	clientConn = tls.Server(clientConn, tlsConfig)

	p.handle(clientConn, true)
}

func (p Server) handle(clientConn net.Conn, isSsl bool) {
	clientWriter := bufio.NewWriter(clientConn)
	clientReader := bufio.NewReader(clientConn)

	r, err := http.ReadRequest(clientReader)
	if err == io.EOF {
		return
	}
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to read request: %v", err))
		return
	}

	originalRequest := r.Clone(r.Context())

	err = applyRules(p.requestRules, r, nil)
	if err != nil {
		slog.Error("apply rules error", slog.String("err", err.Error()), slog.Any("request", r))
		return
	}

	originalRequest.Body = r.Body

	logger := slog.With(slog.String("id", uuid.NewString()), slog.String("url", r.URL.String()), slog.String("method", r.Method))
	logger.Debug("Received request")

	var extConn net.Conn
	if isSsl {
		extConn, err = tls.Dial("tcp", getHost(r.Host, "443"), nil)
	} else {
		extConn, err = net.Dial("tcp", getHost(r.Host, "80"))
	}

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to dial remote host: %v", err))
		return
	}
	defer extConn.Close()

	logger.Debug("Connected to remote host", slog.String("host", r.Host), slog.Bool("isSsl", isSsl))

	extReader := bufio.NewReader(extConn)
	extWriter := bufio.NewWriter(extConn)

	var resp *http.Response
	for {
		err = r.Write(extWriter)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to write request: %v", err))
			return
		}

		logger.Debug("Sent request", slog.Any("request", r))

		err = extWriter.Flush()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to flush request: %v", err))
			return
		}

		logger.Debug("Flushed request")

		resp, err = http.ReadResponse(extReader, r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to read response: %v", err))
			return
		}

		logger.Debug("Received response", slog.Any("response", resp))

		err = applyRules(p.responseRules, originalRequest, resp)
		if err != nil {
			logger.Error("apply rules error", slog.String("err", err.Error()))
			return
		}

		err = resp.Write(clientWriter)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to write response: %v", err))
			return
		}

		logger.Debug("Sent response")

		err = clientWriter.Flush()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to flush response: %v", err))
			return
		}

		logger.Debug("Flushed response")

		if isWebSocket(r) {
			logger.Debug("Upgrading to WebSocket")

			errChan := make(chan error, 2)
			copyConn := func(a, b net.Conn) {
				buf := buf2.ByteGet(BufSize)
				defer buf2.BytePut(buf)
				_, err := io.CopyBuffer(a, b, buf)
				errChan <- err
			}

			go copyConn(clientConn, extConn)
			go copyConn(extConn, clientConn)
			select {
			case err = <-errChan:
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to write response: %v", err))
				}
			}
			return
		}

		if !isKeepAlive(r) {
			logger.Debug("Closing connection")
			return
		}

		r, err = http.ReadRequest(clientReader)
		if err == io.EOF {
			return
		}
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to read request: %v", err))
			return
		}

		err = applyRules(p.requestRules, r, nil)
		if err != nil {
			logger.Error("apply rules error", slog.String("err", err.Error()), slog.Any("request", r))
			return
		}

		logger.Debug("Received request", slog.Any("request", r))
	}
}

func applyRules(rules []*rule.Rule, req *http.Request, resp *http.Response) error {
	for _, r := range rules {
		ok, err := r.Check(req, resp)
		if err != nil {
			return fmt.Errorf(`check rule "%s" error: %v`, r.Name, err)
		}

		if !ok {
			continue
		}

		err = r.Apply(req, resp)
		if err != nil {
			return fmt.Errorf(`apply rule "%s" error: %v`, r.Name, err)
		}

	}

	return nil
}

func isWebSocket(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket"
}

func isKeepAlive(r *http.Request) bool {
	return r.Header.Get("Connection") == "keep-alive" ||
		r.Header.Get("Proxy-Connection") == "keep-alive"
}

func getHost(addr, defaultPort string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = defaultPort
	}

	return net.JoinHostPort(host, port)
}
