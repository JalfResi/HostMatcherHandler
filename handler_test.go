package hostmatcherhandler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestHostMatcherHandler(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	OkHandlerFunc := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	// Create a request to pass to our handler. We dont have any query parameters
	// for now, so we'll pass 'nil' as the third parameter
	req, err := http.NewRequest("GET", "/users/ben", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-User", "ABC123")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := &HostMatcherHandler{}

	handler.HandleFunc(OkHandlerFunc)

	ts := httptest.NewServer(http.HandlerFunc(OkHandlerFunc))
	defer ts.Close()

	handler.AddHost(regexp.MustCompile("X-User: (.*)"), "http://service.example.com/user/$1")

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}
}
