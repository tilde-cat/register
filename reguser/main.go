package main

import (
	"strings"
	"flag"
	"bytes"
	"strconv"
	"fmt"
	"os"
	"os/user"
	"os/exec"
	"encoding/json"
	"io/ioutil"
	"log"
	"github.com/tilde-cat/register"
)

var verbose = flag.Bool("v", false, "verbose")
var dry = flag.Bool("dry", true, "dry run")

func readRequest(path string) (*register.Request, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r register.Request
	err = json.Unmarshal(b, &r)
	return &r, err
}

func shell(cmd string, args... string) {
	var errOut bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stderr = &errOut
	if err := c.Run(); err != nil {
		log.Fatalf("%v failed:\n%v", cmd, errOut.String())
		os.Exit(1)
	}
}

func fixKey(key string) string {
	return strings.Replace(strings.Replace(key, "\n", "", -1), "\r", "", -1)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] file\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	path := flag.Arg(0)
	r, err := readRequest(path)
	if err != nil {
		log.Fatalf("Failed to read request: %v", err)
	}
	if *verbose || *dry {
		log.Printf("Username: '%v'\n", r.Username)
		log.Printf("Email: '%v'\n", r.Email)
		log.Printf("Why:\n%v\n", r.Why)
		log.Printf("SSH key:\n%v\n", r.SSHPublicKey)
		log.Printf("Status: '%v'\n", r.Status)
	}
	if r.Status != "Pending" {
		log.Fatalf("This request is not pending")
	}
	if *dry {
		log.Println("dry run ends")
		return
	}

	shell("adduser", "--disabled-login", "--gecos", "", r.Username)
	authorizedKeysPath := "/home/"+r.Username+"/.ssh/authorized_keys"
	if err = ioutil.WriteFile(authorizedKeysPath, []byte(fixKey(r.SSHPublicKey)), 0664); err != nil {
		log.Fatal("sshkey instalation failed: %v", err)
	}
	user, err := user.Lookup(r.Username)
	if err != nil {
		log.Fatalf("Failed to get user '%v': %v", r.Username, err)
	}
	if *verbose {
		log.Printf("user: %#v\n", user)
	}
	userId, _ := strconv.Atoi(user.Uid)
	groupId, _ := strconv.Atoi(user.Gid)
	if err := os.Chown(authorizedKeysPath, userId, groupId); err != nil {
		log.Fatal("chown failed: %v", err)
	}
	r.Status = "Account created"
	b, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Fatalf("Failed to serialize request: %v\n", err)
	}
	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		log.Fatalf("Failed to save request: %v\n", err)
	}
}
