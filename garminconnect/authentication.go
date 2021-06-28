package garminconnect

import (
	"github.com/abrander/garmin-connect"
	"github.com/bobdoah/subtly-witty-lemur/state"
	"github.com/urfave/cli/v2"
)

func init() {
	state.AuthState.Garmin = connect.NewClient(
		connect.AutoRenewSession(true),
	)
}

// Authenticate with Garmin Connect
func Authenticate(c *cli.Context) error {
	state.LoadState()

	state.AuthState.Garmin.SetOptions(connect.Credentials(
		c.String("email"),
		c.String("password"),
	))
	err := state.AuthState.Garmin.Authenticate()
	if err == nil {
		state.StoreState()
	}
	return err
}

// Signout of Garmin Connect
func Signout(c *cli.Context) error {
	state.LoadState()
	if err := state.AuthState.Garmin.Signout(); err != nil {
		return err
	}
	state.AuthState.Garmin.Password = ""
	return nil
}
