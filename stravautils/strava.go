package stravautils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

// GetActivityNameForTime returns an activity name "Morning, Lunch, Evening, Night" "Ride" for a given time
func GetActivityNameForTime(activityTime time.Time) string {
	switch hour := activityTime.Hour(); {
	case hour >= 4 && hour < 12:
		return "Morning Ride"
	case hour >= 12 && hour < 14:
		return "Lunch Ride"
	case hour >= 14 && hour < 18:
		return "Afternoon Ride"
	case hour >= 18 && hour < 22:
		return "Evening Ride"
	default:
		return "Night Ride"
	}
}

// UploadActivity uploads the activity and sets the gear ID and commute status for an activity
func UploadActivity(accessToken string, activityTime time.Time, activityFilename string, gearID string, isCommute bool) (*int64, error) {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	f, err := os.Open(activityFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", activityFilename, err)
	}

	opts := strava.CreateUploadOpts{
		Name:     optional.NewString(GetActivityNameForTime(activityTime)),
		Type:     optional.NewString("Ride"),
		DataType: optional.NewString("tcx"),
		File:     optional.NewInterface(f),
		GearId:   optional.NewString(gearID),
	}
	if isCommute {
		opts.Commute = optional.NewInt32(1)
	}
	defer f.Close()

	logger.GetLogger().Printf("Uploading %s, gearID: %s, isCommute: %v", activityFilename, gearID, isCommute)
	upload, resp, err := client.UploadsApi.CreateUpload(ctx, &opts)
	for {
		if err != nil {
			var msg string
			if resp != nil {
				body, _ := ioutil.ReadAll(resp.Body)
				msg = string(body)
			}
			return nil, fmt.Errorf("%v %s", err, msg)
		}
		uploadError := upload.Error_
		if uploadError != "" {
			switch {
			case strings.Contains(uploadError, "duplicate"):
				logger.GetLogger().Printf("Skipping duplicate upload: %s", uploadError)
				return nil, nil
			default:
				return nil, fmt.Errorf("Strava upload failed: %s", uploadError)
			}
		}
		if upload.ActivityId != 0 {
			break
		}
		time.Sleep(10 * time.Second)
		logger.GetLogger().Printf(" checking on status of %d...", upload.Id)
		upload, resp, err = client.UploadsApi.GetUploadById(ctx, upload.Id)
	}
	fmt.Printf("Uploaded to https://www.strava.com/activities/%d\n", upload.ActivityId)
	return &upload.ActivityId, nil
}

// SetActivityTypeWalking changes an activity to walking
func SetActivityTypeWalking(accessToken string, activityID int64) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	activityType := strava.WALK_ActivityType
	update := strava.UpdatableActivity{
		Type_: &activityType,
	}

	opts := strava.UpdateActivityByIdOpts{
		Body: optional.NewInterface(update),
	}
	logger.GetLogger().Printf("Setting activity with id %d to walking", activityID)
	_, resp, err := client.ActivitiesApi.UpdateActivityById(ctx, activityID, &opts)
	if err != nil {
		var msg string
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			msg = string(body)
		}
		return fmt.Errorf("%v %s", err, msg)
	}
	return nil
}
