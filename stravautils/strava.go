package stravautils

import (
	"fmt"
	"time"

	"github.com/bobdoah/subtly-witty-lemur/gear"
	"github.com/bobdoah/subtly-witty-lemur/logger"
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

// GetGearIds gets the Ids of the gear names supplied
func GetGearIds(accessToken string, gear *gear.Collection) error {
	client := strava.NewClient(accessToken)
	athlete, err := strava.NewCurrentAthleteService(client).Get().Do()
	if err != nil {
		return err
	}
	logger.GetLogger().Printf("Got bikes: %v", athlete.Weight)
	for _, g := range athlete.Bikes {
		logger.GetLogger().Printf("Checking %s matches", g.Name)
		if g.Name == gear.CommuteBike.Name {
			logger.GetLogger().Printf("%s matched ID %s", gear.CommuteBike.Name, g.Id)
			gear.CommuteBike.StravaID = g.Id
		}
		if g.Name == gear.RoadBike.Name {
			logger.GetLogger().Printf("%s matched ID %s", gear.RoadBike.Name, g.Id)
			gear.RoadBike.StravaID = g.Id
		}
		if g.Name == gear.MountainBike.Name {
			logger.GetLogger().Printf("%s matched ID %s", gear.MountainBike.Name, g.Id)
			gear.MountainBike.StravaID = g.Id
		}
	}
	return nil
}
