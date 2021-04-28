package stravautils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/strava/go.strava"
	"github.com/urfave/cli/v2"
)

const port = 8080

var authenticator *strava.OAuthAuthenticator

// Auth contains the authentication response
var Auth *strava.AuthorizationResponse

// StateFile is the path of the statefile
var StateFile string

//Authenticate completes the OAuth exchange for a given athlete
func Authenticate(c *cli.Context) error {
	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://localhost:%d/exchange_token", port),
		RequestClientGenerator: nil,
	}

	http.HandleFunc("/", indexHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	path, err := authenticator.CallbackPath()
	if err != nil {
		return err
	}
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	// start the server
	fmt.Printf("Visit http://localhost:%d/ to view the demo\n", port)
	fmt.Printf("ctrl-c to exit")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// you should make this a template in your real application
	fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.ActivityReadAll, true))
	fmt.Fprint(w, `<img src="./static/btn_strava_connectwith_orange.png" />`)
	fmt.Fprint(w, `</a>`)
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Successfully authorized")
	Auth = auth
	storeState()

}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}
}

// StateFilename returns the path to the statefile name
func StateFilename() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not detect home directory: %s", err.Error())
	}

	return path.Join(home, ".subtly-witty-lemur.json")
}

// LoadState loads the auth state from the statefile
func LoadState() {
	data, err := ioutil.ReadFile(StateFile)
	if err != nil {
		log.Printf("Could not open state file: %s", err.Error())
		return
	}

	err = json.Unmarshal(data, Auth)
	if err != nil {
		log.Fatalf("Could not unmarshal state: %s", err.Error())
	}
}

func storeState() {
	b, err := json.MarshalIndent(Auth, "", "  ")
	if err != nil {
		log.Fatalf("Could not marshal state: %s", err.Error())
	}

	err = ioutil.WriteFile(StateFile, b, 0600)
	if err != nil {
		log.Fatalf("Could not write state file: %s", err.Error())
	}
}
