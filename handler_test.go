package hostmatcherhandler

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestHostMatcherHandler(t *testing.T) {
	log.SetFlags(log.Lshortfile)

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

	handler.HandleFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Query().Get("user") != "ABC123" {
			t.Errorf("sub-query url query string returned wrong user: got %v expected %v", r.URL.Query(), "user=ABC123")
		}
	}))
	defer ts.Close()

	target := fmt.Sprintf("%s?user=$1", ts.URL)

	handler.AddHost(regexp.MustCompile("X-User: (.*)"), target)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}
}
