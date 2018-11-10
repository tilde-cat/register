package register

type Request struct {
	Username     string
	Email        string
	Why          string
	SSHPublicKey string
	Status       string
}

func (r *Request) IsPending() bool {
    return r.Status == "Pending"
}

func (r *Request) IsValid() bool {
	return r.Username != "" &&
		r.Email != "" &&
		r.Why != "" &&
		r.SSHPublicKey != ""
}
