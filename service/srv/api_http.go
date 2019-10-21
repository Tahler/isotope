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

		for _, step := range h.Service.Script {
			err := s.execute(step)
			if err != nil {
				log.Errf("%s", err)
				makeHTTPResponse(w, r, http.StatusInternalServerError, startTime)
				return
			}
		}
		makeHTTPResponse(w, r, http.StatusOK, startTime)
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
