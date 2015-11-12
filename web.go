package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"io/ioutil"
	"encoding/json"
)

func main() {
	http.HandleFunc("/", statusResponse)

	ip := os.Getenv("OPENSHIFT_GO_IP")
	port := os.Getenv("OPENSHIFT_GO_PORT")

	if ip == "" || port == "" {
		panic("Please set OPENSHIFT_GO_IP and OPENSHIFT_GO_PORT.")
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

func statusResponse(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "<!doctype html>\n")
	fmt.Fprintf(res, "<html>\n")
	fmt.Fprintf(res, "<body>\n")
	fmt.Println(req.Method)

	lockfile := getLockfile()

	switch req.Method {
	case "POST":
		switch req.PostFormValue("operation") {
		case "lock":
			if lockfile.IsLocked {
				fmt.Fprintf(res, "cannot lock, sorry<br>\n")
			} else {
				lockfile.IsLocked = true
				lockfile.HeldBy = req.PostFormValue("locker")
				lockfile.ModificationTime = time.Now().String()
				writeLockfile(lockfile)
			}
		case "release":
			lockfile.IsLocked = false
			lockfile.ModificationTime = time.Now().String()
			writeLockfile(lockfile)
		}
	}

	if lockfile.IsLocked {
		fmt.Fprintf(res, "<b>" + lockfile.HeldBy + "</b> has the lock since <b>" + lockfile.ModificationTime + "</b><br>\n")
		fmt.Fprintf(res, "<form method='POST'><input type='hidden' name='operation' value='release'><input type='submit' value='release lock'></form>")
	} else {
		fmt.Fprintf(res, "unlocked right now, since " + lockfile.ModificationTime + "<br>\n")
		fmt.Fprintf(res, "<form method='POST'><input type='hidden' name='operation' value='lock'><input type='text' name='locker' placeholder='My Glorious Name'><input type='submit' value='grab lock'></form>")
	}
	fmt.Fprintf(res, "</body>\n")
	fmt.Fprintf(res, "</html>\n")
}
