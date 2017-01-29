package hostmatcherhandler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type Getter interface {
	Get(url string) (resp *Response, err error)
}

type hostMatcher struct {
	pattern *regexp.Regexp // e.g. "x-user: (.*)"
	target  string         // e.g. "http://service.example.com/user/$1"
	getter  Getter
}

type HostMatcherHandler struct {
	handler http.Handler
	matches []*hostMatcher
}

func (h *HostMatcherHandler) AddHost(pattern *regexp.Regexp, target string, getter Getter) {
	h.matches = append(h.matches, &hostMatcher{pattern, target, getter})
}

func (h *HostMatcherHandler) Handler(handler http.Handler) {
	h.handler = handler
}

func (h *HostMatcherHandler) HandleFunc(handler func(http.ResponseWriter, *http.Request)) {
	h.handler = http.HandlerFunc(handler)
}

func (h *HostMatcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, match := range h.matches {
		for key, value := range r.Header {
			header := fmt.Sprintf("%s: %s", key, value[0])
			t := match.pattern.ReplaceAllString(header, match.target)

			// how to test this?
			resp, err := match.getter.Get(t)
			if err != nil {
				log.Fatal(err)
			}

			log.Println(resp)
		}
	}

	h.handler.ServeHTTP(w, r)
	return
}
