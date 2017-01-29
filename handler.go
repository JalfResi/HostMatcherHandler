package hostmatcherhandler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
)

type hostMatcher struct {
	pattern *regexp.Regexp // e.g. "x-user: (.*)"
	target  string         // e.g. "http://service.example.com/user/$1"
}

type HostMatcherHandler struct {
	handler http.Handler
	matches []*hostMatcher
}

func (h *HostMatcherHandler) AddHost(pattern *regexp.Regexp, target string) {
	h.matches = append(h.matches, &hostMatcher{pattern, target})
}

func (h *HostMatcherHandler) Handler(handler http.Handler) {
	h.handler = handler
}

func (h *HostMatcherHandler) HandleFunc(handler func(http.ResponseWriter, *http.Request)) {
	h.handler = http.HandlerFunc(handler)
}

func (h *HostMatcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	for _, match := range h.matches {
		for key, value := range r.Header {
			header := fmt.Sprintf("%s: %s", key, value[0])
			t := match.pattern.ReplaceAllString(header, match.target)
			log.Println("rad")

			go func() {
				wg.Add(1)
				defer wg.Done()

				log.Println("sweet")

				resp, err := http.Get(t)
				if err != nil {
					log.Fatal("ERROR: ", err)
				}

				log.Println("sweeter")
				log.Println("cat", t, r, resp)
			}()
		}
	}

	wg.Wait()
	log.Println("here")

	h.handler.ServeHTTP(w, r)
	return
}
