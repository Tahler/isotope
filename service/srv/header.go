package srv

import (
	"net/http"
)

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
