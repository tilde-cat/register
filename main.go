package register

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
