package srv

import context "context"

type Server struct {
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Stop() error {
	return nil
}

// GRPC ping
func (s *Server) Ping(c context.Context, in *PingMessage) (*PingMessage, error) {
	return in, nil
}
