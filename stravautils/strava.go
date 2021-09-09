package stravautils

import (
	"context"
	"fmt"
	"time"

	"github.com/antihax/optional"
	"github.com/vangent/strava"

	"github.com/bobdoah/subtly-witty-lemur/gear"
	"github.com/bobdoah/subtly-witty-lemur/logger"
)

// GetActivityForTime returns activities within a 20 minute window of a given time
func GetActivityForTime(accessToken string, startTime time.Time) ([]strava.SummaryActivity, error) {

	afterTime := startTime.Add(time.Minute * -10)
	beforeTime := startTime.Add(time.Minute * 10)

	fmt.Printf("Checking Strava for activities between %v and %v\n", beforeTime, afterTime)

	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)
	opts := &strava.GetLoggedInAthleteActivitiesOpts{
		Before: optional.NewInt32(int32(beforeTime.Unix())),
		After:  optional.NewInt32(int32(afterTime.Unix())),
	}

	activities, _, err := client.ActivitiesApi.GetLoggedInAthleteActivities(ctx, opts)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// GetGearIds gets the Ids of the gear names supplied
func GetGearIds(accessToken string, gear *gear.Collection) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	athlete, _, err := client.AthletesApi.GetLoggedInAthlete(ctx)
	if err != nil {
		return err
	}
	logger.GetLogger().Printf("Got bikes: %v", athlete.Bikes)
	for _, g := range athlete.Bikes {
		logger.GetLogger().Printf("Checking %s matches supplied bikes", g.Name)
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
