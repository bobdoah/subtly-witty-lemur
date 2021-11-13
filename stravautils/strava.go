package stravautils

import (
	"context"
	"errors"
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

// WaitForActivity waits for activity to sync in Strava
func WaitForActivity(accessToken string, startTime time.Time) ([]strava.SummaryActivity, error) {
	c := make(chan []strava.SummaryActivity)
	go func() {
		var summaryActivity []strava.SummaryActivity
		for {
			summaryActivity, _ = GetActivityForTime(accessToken, startTime)
			if summaryActivity != nil {
				break
			}
			time.Sleep(10 * time.Second)
		}
		c <- summaryActivity
	}()
	select {
	case res := <-c:
		return res, nil
	case <-time.After(5 * time.Minute):
		return nil, errors.New("Timed out waiting for activity")
	}
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

// UpdateActivity sets the gear ID and commute status for an activity
func UpdateActivity(accessToken string, summaryActivity strava.SummaryActivity, gearID string, isCommute bool) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	update := strava.UpdatableActivity{
		Name:        summaryActivity.Name,
		Type_:       summaryActivity.Type_,
		WorkoutType: int(summaryActivity.WorkoutType),
		GearId:      gearID,
		Commute:     isCommute,
		Trainer:     false,
	}
	logger.GetLogger().Printf("Updating activity with id %s to gearID %s and isCommute %s", summaryActivity.Id, gearID, isCommute)
	_, _, err := client.ActivitiesApi.UpdateActivityById(ctx, summaryActivity.Id, &strava.UpdateActivityByIdOpts{Body: optional.NewInterface(update)})
	if err != nil {
		logger.GetLogger().Printf("Failed to update activity with id %d, with: %s", summaryActivity.Id, err)
	} else {
		logger.GetLogger().Printf("Successfully updated activity with id %d, gearID %s and isCommute: %s", summaryActivity.Id, gearID, isCommute)
	}
	return err
}
