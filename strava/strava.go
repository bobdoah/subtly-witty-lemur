package strava

import (
	"time"

	"github.com/strava/go.strava"
)

func getActivityForTime(accessToken string, startTime time.Time) ([]strava.ActivitySummary, error) {

	afterTime := startTime.Add(time.Minute * -10).Unix()
	beforeTime := startTime.Add(time.Minute * 10).Unix()

	client := strava.NewClient(accessToken)
	activities, err := strava.NewCurrentAthleteService(client).ListActivities().Before(beforeTime).After(afterTime).Do()
	if err != nil {
		return activities, err
	}
}
