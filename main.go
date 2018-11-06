package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/satori/go.uuid"
)

const (
	FormUrl                = "/"
	FormPostUrl            = "/post"
	RequestStatusUrlPrefix = "/status/"
	ErrorUrl               = "/error"
)

var statusRE = regexp.MustCompile(RequestStatusUrlPrefix + `(.+)$`)

type Id uuid.UUID

func (id Id) String() string {
	return uuid.UUID(id).String()
}

type Request struct {
	Username     string
	Email        string
	Why          string
	SSHPublicKey string
	Status       string
}

func (r *Request) IsValid() bool {
	return r.Username != "" &&
		r.Email != "" &&
		r.Why != "" &&
		r.SSHPublicKey != ""
}

type Io interface {
	Save(r Request) (Id, error)
	Load(id Id) (*Request, error)
}

type FsIo struct {
}

func (io *FsIo) Save(r Request) (Id, error) {
	b, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return Id{}, err
	}
	id := Id(uuid.NewV4())
	return id, ioutil.WriteFile(id.String()+".json", b, 0600)
}

func (io *FsIo) Load(id Id) (*Request, error) {
	b, err := ioutil.ReadFile(id.String() + ".json")
	if err != nil {
		return nil, err
	}
	var req Request
	if err := json.Unmarshal(b, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

type Server struct {
	Io Io
}

func (s *Server) RequestPage(w http.ResponseWriter, r *http.Request) {
	m := statusRE.FindStringSubmatch(r.URL.String())
	if len(m) != 2 {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}
	uid := m[1]
	id, err := uuid.FromString(uid)
	if err != nil {
		http.Error(w, "no such request: '"+uid+"'", http.StatusBadRequest)
		return
	}
	req, err := s.Io.Load(Id(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Status: %v", req.Status)
}

func (s *Server) IncorrectRequest(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) FormPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	req := Request{}
	req.Username = r.PostFormValue("username")
	req.Email = r.PostFormValue("email")
	req.Why = r.PostFormValue("why")
	req.SSHPublicKey = r.PostFormValue("sshpublickey")
	req.Status = "Pending"
	if !req.IsValid() {
		log.Println("Invalid request", r.PostForm)
		http.Redirect(w, r, ErrorUrl, http.StatusSeeOther)
		return
	}
	id, err := s.Io.Save(req)
	log.Println("Valid request", r.PostForm, err)
	http.Redirect(w, r, RequestStatusUrlPrefix+id.String(), http.StatusSeeOther)
}

func (s *Server) FormPage(w http.ResponseWriter, r *http.Request) {
	formTemplate.Execute(w, nil)
}

func main() {
	var io FsIo
	server := Server{Io: &io}
	http.HandleFunc(RequestStatusUrlPrefix, server.RequestPage)
	http.HandleFunc(ErrorUrl, server.IncorrectRequest)
	http.HandleFunc(FormPostUrl, server.FormPostHandler)
	http.HandleFunc(FormUrl, server.FormPage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
