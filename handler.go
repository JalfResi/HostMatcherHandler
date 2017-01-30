package hostmatcherhandler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync"
)

type hostMatcher struct {
	pattern *regexp.Regexp // e.g. "x-user: (.*)"
	target  string         // e.g. "http://service.example.com/user/$1"
	content []byte
}

func (h *hostMatcher) Fetch(target string) {
	resp, err := http.Get(target)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
	defer resp.Body.Close()

	h.content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}

	//log.Printf("Fetched data %s from %s\n", string(h.content), target)
}

type HostMatcherHandler struct {
	handler http.Handler
	matches []*hostMatcher
}

func (h *HostMatcherHandler) AddHost(pattern *regexp.Regexp, target string) {
	h.matches = append(h.matches, &hostMatcher{
		pattern: pattern,
		target:  target,
	})
}

func (h *HostMatcherHandler) Handler(handler http.Handler) {
	h.handler = handler
}

func (h *HostMatcherHandler) HandleFunc(handler func(http.ResponseWriter, *http.Request)) {
	h.handler = http.HandlerFunc(handler)
}

func (h *HostMatcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// this wait group is used to make sure all the sub-requests have completed first
	// this is so we can run all the sub requests concurrently
	var wg sync.WaitGroup

	// Loop over each header match
	for _, match := range h.matches {
		for key, value := range r.Header {
			header := fmt.Sprintf("%s: %s", key, value[0])

			t := match.pattern.ReplaceAllString(header, match.target)

			// start our sub-request
			wg.Add(1)
			go func() {
				defer wg.Done()
				match.Fetch(t)
			}()
		}
	}

	// Capture the response
	rec := httptest.NewRecorder()

	// Make the original proxy request
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.handler.ServeHTTP(rec, r)
	}()

	// wait for all requests to finish
	wg.Wait()

	// we copy the captured response headers to our new response
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}

	// grab the captured response
	originalData := rec.Body.Bytes()

	// Modify the response body here
	replace := regexp.MustCompile("(\"ABC123\")")
	data := replace.ReplaceAllString(string(originalData), string(h.matches[0].content))

	// But the Content-Length might have been set already,
	// we should modify it by adding the length
	// of our own data.
	// Ignoring the error is fine here:
	// if Content-Length is empty or otherwise invalid,
	// Atoi() will return zero,
	// which is just what we'd want in that case.
	clen, _ := strconv.Atoi(r.Header.Get("Content-Length"))
	clen += len(data)
	r.Header.Set("Content-Length", strconv.Itoa(clen))

	// write out our modified response
	w.Write([]byte(data))
}
