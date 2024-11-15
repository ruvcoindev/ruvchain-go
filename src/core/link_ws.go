package core

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Arceliar/phony"
	"github.com/coder/websocket"
)

type linkWS struct {
	phony.Inbox
	*links
	listenconfig *net.ListenConfig
}

type linkWSConn struct {
	net.Conn
}

type linkWSListener struct {
	ch         chan *linkWSConn
	ctx        context.Context
	httpServer *http.Server
	listener   net.Listener
}

type wsServer struct {
	ch  chan *linkWSConn
	ctx context.Context
}

func (l *linkWSListener) Accept() (net.Conn, error) {
	qs := <-l.ch
	if qs == nil {
		return nil, context.Canceled
	}
	return qs, nil
}

func (l *linkWSListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *linkWSListener) Close() error {
	if err := l.httpServer.Shutdown(l.ctx); err != nil {
		return err
	}
	return l.listener.Close()
}

func (s *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" || r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"ruv-ws"},
	})
	if err != nil {
		return
	}

	if c.Subprotocol() != "ruv-ws" {
		c.Close(websocket.StatusPolicyViolation, "client must speak the ruv-ws subprotocol")
		return
	}

	s.ch <- &linkWSConn{
		Conn: websocket.NetConn(s.ctx, c, websocket.MessageBinary),
	}
}

func (l *links) newLinkWS() *linkWS {
	lt := &linkWS{
		links: l,
		listenconfig: &net.ListenConfig{
			KeepAlive: -1,
		},
	}
	return lt
}

func (l *linkWS) dial(ctx context.Context, url *url.URL, info linkInfo, options linkOptions) (net.Conn, error) {
	if options.tlsSNI != "" {
		return nil, ErrLinkSNINotSupported
	}
	wsconn, _, err := websocket.Dial(ctx, url.String(), &websocket.DialOptions{
		Subprotocols: []string{"ruv-ws"},
	})
	if err != nil {
		return nil, err
	}
	return &linkWSConn{
		Conn: websocket.NetConn(ctx, wsconn, websocket.MessageBinary),
	}, nil
}

func (l *linkWS) listen(ctx context.Context, url *url.URL, _ string) (net.Listener, error) {
	nl, err := l.listenconfig.Listen(ctx, "tcp", url.Host)
	if err != nil {
		return nil, err
	}

	ch := make(chan *linkWSConn)

	httpServer := &http.Server{
		Handler: &wsServer{
			ch:  ch,
			ctx: ctx,
		},
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	lwl := &linkWSListener{
		ch:         ch,
		ctx:        ctx,
		httpServer: httpServer,
		listener:   nl,
	}
	go lwl.httpServer.Serve(nl) // nolint:errcheck
	return lwl, nil
}
