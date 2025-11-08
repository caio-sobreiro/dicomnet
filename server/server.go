package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/pdu"
)

// Option configures a Server instance.
type Option func(*Server)

// WithLogger overrides the logger used by the server.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.Logger = logger
	}
}

// Server exposes a reusable DICOM listener that wires the DIMSE and PDU layers.
type Server struct {
	AETitle string
	Handler interfaces.ServiceHandler
	Logger  *slog.Logger
}

// New builds a Server with the provided AE title and handler.
func New(aeTitle string, handler interfaces.ServiceHandler, opts ...Option) *Server {
	srv := &Server{AETitle: aeTitle, Handler: handler}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// ListenAndServe listens on the given address and serves until the context is done or an error occurs.
func ListenAndServe(ctx context.Context, address, aeTitle string, handler interfaces.ServiceHandler, opts ...Option) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	srv := New(aeTitle, handler, opts...)
	return srv.Serve(ctx, listener)
}

// Serve accepts connections from listener until ctx is cancelled or an unrecoverable error occurs.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	if listener == nil {
		return errors.New("dicomserver: listener is required")
	}
	if s == nil {
		return errors.New("dicomserver: server is nil")
	}
	if s.Handler == nil {
		return errors.New("dicomserver: handler is required")
	}
	if s.AETitle == "" {
		return errors.New("dicomserver: AE title is required")
	}

	logger := s.logger()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	logger.Info("DICOM server listening",
		"address", listener.Addr().String(),
		"ae_title", s.AETitle)

	var (
		wg       sync.WaitGroup
		serveErr error
	)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				break
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				logger.Warn("Accept timeout", "error", err)
				continue
			}
			serveErr = err
			break
		}

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			s.handleConnection(ctx, c, logger)
		}(conn)
	}

	wg.Wait()

	if serveErr != nil {
		return serveErr
	}

	return ctx.Err()
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn, logger *slog.Logger) {
	logger.Info("Accepted DICOM connection",
		"remote_addr", conn.RemoteAddr())

	adapter := &dimseHandlerAdapter{service: dimse.NewService(s.Handler, logger)}
	layer := pdu.NewLayer(conn, adapter, s.AETitle, logger)

	if err := layer.HandleConnection(); err != nil && ctx.Err() == nil {
		logger.Warn("DIMSE connection ended",
			"error", err,
			"remote_addr", conn.RemoteAddr())
	} else {
		logger.Info("DIMSE connection closed",
			"remote_addr", conn.RemoteAddr())
	}
}

func (s *Server) logger() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

type dimseHandlerAdapter struct {
	service *dimse.Service
}

func (a *dimseHandlerAdapter) HandleDIMSEMessage(presContextID byte, msgCtrlHeader byte, data []byte, layer *pdu.Layer) error {
	return a.service.HandleDIMSEMessage(presContextID, msgCtrlHeader, data, layer)
}
