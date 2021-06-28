package state

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/abrander/garmin-connect"
	"github.com/strava/go.strava"
)

// GearList holds the details for the gear
type GearList struct {
	CommuteBike  gear
	MountainBike gear
	RoadBike     gear
}

// Gear holds the strings for the names of the gear
type gear struct {
	Name       string
	GarminUUID string
	StravaID   string
}

// StateFile is the path of the statefile
var StateFile string

type authState struct {
	Strava *strava.AuthorizationResponse `json:"strava"`
	Garmin *connect.Client               `json:"garmin"`
}

// AuthState holds auth details for Garmin Connect and Strava
var AuthState authState

// Filename returns the path to the statefile name
func Filename() string {
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

	err = json.Unmarshal(data, &AuthState)
	if err != nil {
		log.Fatalf("Could not unmarshal state: %s", err.Error())
	}
}

// StoreState stores the auth state into the statefile
func StoreState() {
	b, err := json.MarshalIndent(AuthState, "", "  ")
	if err != nil {
		log.Fatalf("Could not marshal state: %s", err.Error())
	}

	err = ioutil.WriteFile(StateFile, b, 0600)
	if err != nil {
		log.Fatalf("Could not write state file: %s", err.Error())
	}
}
