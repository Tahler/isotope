package srv

import (
	"net/http"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/srv/prometheus"
	"istio.io/fortio/log"
)

// pathTracesHeaderKey is the HTTP header key for path tracing. It must be in
// Train-Case.
const pathTracesHeaderKey = "Path-Traces"

var (
	forwardableHeaders = []string{
		"X-Request-Id",
		"X-B3-Traceid",
		"X-B3-Spanid",
		"X-B3-Parentspanid",
		"X-B3-Sampled",
		"X-B3-Flags",
		"X-Ot-Span-Context",
	}
	forwardableHeadersSet = make(map[string]bool, len(forwardableHeaders))
)

type Handler struct {
	Service      svc.Service
	ServiceTypes map[string]svctype.ServiceType
}

func newApiHttp(h Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ServiceHandler(h))
	mux.Handle("/metrics", prometheus.Handler())
	return mux
}

func ServiceHandler(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		prometheus.RecordRequestReceived()

		respond := func(status int) {
			w.WriteHeader(status)
			err := r.Write(w)
			if err != nil {
				log.Errf("%s", err)
			}

			stopTime := time.Now()
			duration := stopTime.Sub(startTime)
			// TODO: Record size of response payload.
			prometheus.RecordResponseSent(duration, 0, status)
		}

		for _, step := range h.Service.Script {
			forwardableHeader := extractForwardableHeader(r.Header)
			err := execute(step, forwardableHeader, h.ServiceTypes)
			if err != nil {
				log.Errf("%s", err)
				respond(http.StatusInternalServerError)
				return
			}
		}

		respond(http.StatusOK)
	}
}

func init() {
	for _, key := range forwardableHeaders {
		forwardableHeadersSet[key] = true
	}
}

func extractForwardableHeader(header http.Header) http.Header {
	forwardableHeader := make(http.Header, len(forwardableHeaders))
	for key := range forwardableHeadersSet {
		if values, ok := header[key]; ok {
			forwardableHeader[key] = values
		}
	}
	return forwardableHeader
}
