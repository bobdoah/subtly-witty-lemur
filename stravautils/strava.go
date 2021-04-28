package stravautils

import (
	"fmt"
	"time"

	"github.com/strava/go.strava"
)

// GetActivityForTime returns activities within a 20 minute window of a given time
func GetActivityForTime(accessToken string, startTime time.Time) ([]*strava.ActivitySummary, error) {

	afterTime := startTime.Add(time.Minute * -10)
	beforeTime := startTime.Add(time.Minute * 10)

	fmt.Printf("Checking Strava for activities between %v and %v\n", beforeTime, afterTime)

	client := strava.NewClient(accessToken)
	activities, err := strava.NewCurrentAthleteService(client).ListActivities().Before(int(beforeTime.Unix())).After(int(afterTime.Unix())).Do()
	if err != nil {
		return nil, err
	}
	return activities, nil
}
