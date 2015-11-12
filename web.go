package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"io/ioutil"
	"encoding/json"
	"html/template"
)

func main() {
	http.HandleFunc("/", statusResponse)

	ip := os.Getenv("OPENSHIFT_GO_IP")
	port := os.Getenv("OPENSHIFT_GO_PORT")

	if ip == "" || port == "" {
		panic("Please set OPENSHIFT_GO_IP and OPENSHIFT_GO_PORT.")
	}

	if os.Getenv("OPENSHIFT_DATA_DIR") == "" {
		panic("Please use OPENSHIFT_DATA_DIR.")
	}

	bind := fmt.Sprintf("%s:%s", ip, port)
	fmt.Printf("listening on %s...", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		panic(err)
	}
}

type Lockfile struct {
	IsLocked bool `json:"is_locked"`
	HeldBy string `json:"locker"`
	ModificationTime string `json:"modification_time"`
}

func getLockfilePath() string {
	return os.Getenv("OPENSHIFT_DATA_DIR") + "/lock";
}

func getLockfile() Lockfile {
	path := getLockfilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Lockfile{
			IsLocked: false,
			HeldBy: "",
			ModificationTime: "",
		};
	}

	chars, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var lockfile Lockfile
	if err := json.Unmarshal(chars, &lockfile); err != nil {
		panic(err)
	}
	return lockfile
}

func writeLockfile(lockfile Lockfile) {
	path := getLockfilePath()
	bytes, err := json.Marshal(&lockfile)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(path, bytes, 0600); err != nil {
		panic(err)
	}
}

func currentTimeString() string {
	pragueTime, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		panic(err)
	}
	return time.Now().In(pragueTime).String()
}

func isValidLocker(locker string) bool {
	return len(locker) >= 1 && len(locker) <= 50;
}

func statusResponse(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "<!doctype html>\n")
	fmt.Fprintf(res, "<html>\n")
	fmt.Fprintf(res, "<body>\n")

	lockfile := getLockfile()

	switch req.Method {
	case "POST":
		switch req.PostFormValue("operation") {
		case "lock":
			if lockfile.IsLocked {
				fmt.Fprintf(res, "cannot lock, sorry<br>\n")
			} else {
				locker := req.PostFormValue("locker")
				if isValidLocker(locker) {
					lockfile.IsLocked = true
					lockfile.HeldBy = req.PostFormValue("locker")
					lockfile.ModificationTime = currentTimeString()
					writeLockfile(lockfile)
				} else {
					fmt.Fprintf(res, "you have terrible taste in names<br>\n")
				}
			}
		case "release":
			lockfile.IsLocked = false
			lockfile.ModificationTime = time.Now().String()
			writeLockfile(lockfile)
		}
	}

	template, err := template.New("locked_greeting").Parse(
		`{{define "LockStatusPage"}}
			{{if .IsLocked}}
				<b>{{.HeldBy}}</b> has the lock since <b>{{.ModificationTime}}</b><br>
				<form method='POST'>
					<input type='hidden' name='operation' value='release'>
					<input type='submit' value='release lock'>
				</form>
			{{else}}
				unlocked right now, since <b>{{.ModificationTime}}</b><br>
				<form method='POST'>
					<input type='hidden' name='operation' value='lock'>
					<input type='text' name='locker' placeholder='My Glorious Name'>
					<input type='submit' value='grab lock'>
				</form>
			{{end}}
		{{end}}`)
	if err != nil {
		panic(err)
	}
	if err := template.ExecuteTemplate(res, "LockStatusPage", lockfile); err != nil {
		panic(err)
	}
	fmt.Fprintf(res, "</body>\n")
	fmt.Fprintf(res, "</html>\n")
}
