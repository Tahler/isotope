package srv

import (
	"net/http"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/srv/prometheus"
	"github.com/fortio/fortio/log"
)

type Handler struct {
	Service      svc.Service
	ServiceTypes map[string]svctype.ServiceType
}

func (s *Server) newApiHttp(h Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.ServiceHandler(h))
	mux.Handle("/metrics", prometheus.Handler())
	return mux
}

func (s *Server) ServiceHandler(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		prometheus.RecordRequestReceived()

		var statusCode int
		err := s.executeTasks(s.tasks)
		if err != nil {
			log.Errf("%s", err)
			statusCode = http.StatusInternalServerError
			makeHTTPResponse(w, r, statusCode, startTime)
			return
		}
		statusCode = http.StatusOK
		makeHTTPResponse(w, r, statusCode, startTime)

	}
}

func makeHTTPResponse(w http.ResponseWriter, r *http.Request, statusCode int, startTime time.Time) {
	w.WriteHeader(statusCode)
	err := r.Write(w)
	if err != nil {
		log.Errf("%s", err)
	}

	stopTime := time.Now()
	duration := stopTime.Sub(startTime)
	// TODO: Record size of response payload.
	prometheus.RecordResponseSent(duration, 0, statusCode)
}
