package main

import (
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
)

func main() {
	http.HandleFunc("/", statusResponse)
	bind := fmt.Sprintf("%s:%s", os.Getenv("OPENSHIFT_GO_IP"), os.Getenv("OPENSHIFT_GO_PORT"))
	fmt.Printf("listening on %s...", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		panic(err)
	}
}

func getLockfilePath() string {
	return os.Getenv("OPENSHIFT_DATA_DIR") + "/lock";
}

// Returns empty string if not locked.
func getLocker() string {
	path := getLockfilePath();
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create new lockfile (by releasing the lock).
		releaseLock()
	}

	chars, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(chars)
}

func grabLock(locker string) {
	if err := ioutil.WriteFile(getLockfilePath(), []byte(locker), 0); err != nil {
		panic(err)
	}
}

func releaseLock() {
	if err := ioutil.WriteFile(getLockfilePath(), []byte(""), 0); err != nil {
		panic(err);
	}
}

func statusResponse(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "PUT":
		if getLocker() != "" {
			fmt.Fprintf(res, "cannot lock, sorry<br>")
		} else {
			grabLock(req.PostFormValue("locker"))
		}
	case "DELETE":
		releaseLock()
	}

	locker := getLocker()
	if locker == "" {
		fmt.Fprintf(res, "unlocked right now<br>")
		fmt.Fprintf(res, "<form action='/lock' method='put'><input type='text' name='locker' placeholder='My Glorious Name'><input type='submit' value='grab lock'></form>")
	} else {
		fmt.Fprintf(res, "<b>" + locker + "</b> has the lock since<br>")
		fmt.Fprintf(res, "<form action='/lock' method='delete'><input type='submit' value='release lock'></form>")
	}
}
