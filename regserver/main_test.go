package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/tilde-cat/register"
)

var expected = register.Request{
	Username:     "name",
	Email:        "test@example.com",
	Why:          "foo bar baz",
	SSHPublicKey: "123",
	Status:       "Pending",
}

type ReqEntry struct {
	Request register.Request
	Id      Id
}

type IoStub struct {
	Saved []ReqEntry
	Loads []Id
}

func NewIoStub() *IoStub {
	return &IoStub{}
}

func (io *IoStub) Save(r register.Request) (Id, error) {
	id := Id(uuid.NewV4())
	io.Saved = append(io.Saved, ReqEntry{r, id})
	return id, nil
}

func (io *IoStub) Load(id Id) (*register.Request, error) {
	io.Loads = append(io.Loads, id)
	for _, r := range io.Saved {
		if r.Id == id {
			return &r.Request, nil
		}
	}
	return nil, fmt.Errorf("Missing Request for id: %v", id)
}

func requestForm(target string, values map[string]string) *http.Request {
	r := httptest.NewRequest("POST", target, nil)
	r.PostForm = url.Values{}
	for k, v := range values {
		r.PostForm.Set(k, v)
	}
	return r
}

func TestRequestSaveAfterCorrectFormPost(t *testing.T) {
	io := NewIoStub()
	server := Server{Io: io}
	req := requestForm(FormPostUrl, map[string]string{
		"username":     expected.Username,
		"email":        expected.Email,
		"why":          expected.Why,
		"sshpublickey": expected.SSHPublicKey,
	})
	recorder := httptest.NewRecorder()
	server.FormPostHandler(recorder, req)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("Expected status %v, got: %v", http.StatusSeeOther, resp.StatusCode)
	}
	expectedLoc := RequestStatusUrlPrefix + io.Saved[0].Id.String()
	if loc := resp.Header.Get("Location"); loc != expectedLoc {
		t.Fatalf("Expected location '%v', got '%v'", expectedLoc, loc)
	}
	if !reflect.DeepEqual(expected, io.Saved[0].Request) {
		t.Fatalf("\nExpected '%#v'\n     got '%#v'", expected, io.Saved[0])
	}
}

func TestRedirectToFailureWhenAnyRequestFieldIsEmtpy(t *testing.T) {
	data := []register.Request{
		{Username: "", Email: expected.Email, Why: expected.Why, SSHPublicKey: expected.SSHPublicKey},
		{Username: expected.Username, Email: "", Why: expected.Why, SSHPublicKey: expected.SSHPublicKey},
		{Username: expected.Username, Email: expected.Email, Why: "", SSHPublicKey: expected.SSHPublicKey},
		{Username: expected.Username, Email: expected.Email, Why: expected.Why, SSHPublicKey: ""},
	}
	for _, r := range data {
		io := NewIoStub()
		server := Server{Io: io}
		req := requestForm(FormPostUrl, map[string]string{
			"username":     r.Username,
			"email":        r.Email,
			"why":          r.Why,
			"sshpublickey": r.SSHPublicKey,
		})
		recorder := httptest.NewRecorder()
		server.FormPostHandler(recorder, req)
		resp := recorder.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("Expected see other status, got: %v", resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != ErrorUrl {
			t.Fatalf("Expected location %v, got: %v", ErrorUrl, loc)
		}
	}
}

func TestStatusPageOk(t *testing.T) {
	io := NewIoStub()
	server := Server{Io: io}
	id, _ := io.Save(expected)
	req := httptest.NewRequest("GET", RequestStatusUrlPrefix+id.String(), nil)
	rec := httptest.NewRecorder()
	server.RequestPage(rec, req)
	if io.Loads[0] != id {
		t.Fatalf("Expected load of %v, loaded %v instead", id, io.Loads[0])
	}
}

func TestStatusPageUnknownId(t *testing.T) {
	io := NewIoStub()
	server := Server{Io: io}
	id := Id(uuid.NewV4())
	req := httptest.NewRequest("GET", RequestStatusUrlPrefix+id.String(), nil)
	rec := httptest.NewRecorder()
	server.RequestPage(rec, req)
	if io.Loads[0] != id {
		t.Fatalf("Expected load of %v, loaded %v instead", id, io.Loads[0])
	}
}

func TestStatusPageMalformedId(t *testing.T) {
	io := NewIoStub()
	server := Server{Io: io}
	id := Id(uuid.NewV4())
	req := httptest.NewRequest("GET", RequestStatusUrlPrefix+id.String()+"abc", nil)
	rec := httptest.NewRecorder()
	server.RequestPage(rec, req)
	if l := len(io.Loads); l != 0 {
		t.Fatalf("Expected zero loads, got %v", l)
	}
}

func TestStatusPageMissingId(t *testing.T) {
	io := NewIoStub()
	server := Server{Io: io}
	req := httptest.NewRequest("GET", RequestStatusUrlPrefix, nil)
	rec := httptest.NewRecorder()
	server.RequestPage(rec, req)
	if l := len(io.Loads); l != 0 {
		t.Fatalf("Expected zero loads, got %v", l)
	}
}
