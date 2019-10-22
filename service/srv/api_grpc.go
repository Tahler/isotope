package srv

import (
	"context"
	"net/http"
	"time"

	"github.com/Tahler/isotope/service/srv/prometheus"
	"github.com/fortio/fortio/log"
)

// Ping checks the service graph to call its dependencies, and waits for their responses.
// It also records the execution duration.
func (s *Server) Ping(c context.Context, in *PingMessage) (*PingMessage, error) {
	startTime := time.Now()
	prometheus.RecordRequestReceived()

	var statusCode int
	err := s.executeTasks(s.tasks)
	if err != nil {
		log.Errf("%s", err)
		statusCode = http.StatusInternalServerError
		registerExecutionTime(statusCode, startTime)
		return nil, err
	}

	statusCode = http.StatusOK
	registerExecutionTime(statusCode, startTime)

	return in, nil
}

// Similar to api_http > makeHTTPResponse.
func registerExecutionTime(statusCode int, startTime time.Time) {
	stopTime := time.Now()
	duration := stopTime.Sub(startTime)
	prometheus.RecordResponseSent(duration, 0, statusCode)
}
