package strava

import (
	"time"

	"github.com/strava/go.strava"
)

// GetActivityForTime returns activities within a 20 minute window of a given time
func GetActivityForTime(accessToken string, startTime time.Time) ([]*strava.ActivitySummary, error) {

	afterTime := startTime.Add(time.Minute * -10).Unix()
	beforeTime := startTime.Add(time.Minute * 10).Unix()

	client := strava.NewClient(accessToken)
	activities, err := strava.NewCurrentAthleteService(client).ListActivities().Before(int(beforeTime)).After(int(afterTime)).Do()
	if err != nil {
		return nil, err
	}
	return activities, nil
}
