package stravautils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli/v2"

	"github.com/bobdoah/subtly-witty-lemur/state"
)

//Authenticate completes the OAuth exchange for a given athlete
func Authenticate(c *cli.Context) error {
	state.LoadState()

	port := c.Int("port")
	clientID := c.String("client-id")
	clientSecret := c.String("client-secret")
	ch := make(chan *authResult)
	go responseHandler(ch, port)

	u, _ := url.Parse("https://www.strava.com/oauth/authorize")
	q := u.Query()
	q.Add("client_id", clientID)
	q.Add("redirect_uri", fmt.Sprintf("http://127.0.0.1:%d", port))
	q.Add("response_type", "code")
	q.Add("scope", "profile:read_all,activity:read_all,activity:write")
	u.RawQuery = q.Encode()
	urlstr := u.String()

	fmt.Printf("Pointing your browser to %s. If it doesn't work, please copy the URL and paste it into your browser.\n", urlstr)
	if err := open.Start(urlstr); err != nil {
		return err
	}

	// Wait for the redirect.
	res := <-ch
	if res.err != nil {
		return res.err
	}
	log.Printf("got code %s", res.code)

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", res.code)
	resp, err := http.PostForm("https://www.strava.com/oauth/token", form)
	if err != nil {
		return fmt.Errorf("authentication failed at POST: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed at POST, status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("POST body: %s", string(body))
	m := map[string]interface{}{}
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("authentication failed, POST response was not JSON: %v", err)
	}
	accessToken := m["access_token"].(string)
	if accessToken == "" {
		return fmt.Errorf("authentication failed, no access token received: %v", m)
	}
	fmt.Printf("Your Strava access token is: %s\n", accessToken)
	state.AuthState.StravaAccessToken = accessToken
	state.StoreState()
	return nil
}

type authResult struct {
	err  error
	code string
}

func responseHandler(ch chan *authResult, port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "You can now close this window.")

		res := &authResult{}
		defer func() { ch <- res }()

		q := r.URL.Query()
		if err := q.Get("error"); err != "" {
			res.err = fmt.Errorf("authorization failed: %s", err)
			return
		}
		res.code = q.Get("code")
		if res.code == "" {
			res.err = fmt.Errorf("authorization didn't include a code: %q", r.URL.String())
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
