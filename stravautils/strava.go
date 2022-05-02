package stravautils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
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
func GetActivityNameForTime(activityTime time.Time, isWalk bool, isRun bool) string {
	suffix := "Ride"
	if isWalk {
		suffix = "Walk"
	} else if isRun {
		suffix = "Run"
	}
	var prefix string
	switch hour := activityTime.Hour(); {
	case hour >= 4 && hour < 12:
		prefix = "Morning"
	case hour >= 12 && hour < 14:
		prefix = "Lunch Ride"
	case hour >= 14 && hour < 18:
		prefix = "Afternoon Ride"
	case hour >= 18 && hour < 22:
		prefix = "Evening Ride"
	default:
		prefix = "Night Ride"
	}
	return fmt.Sprintf("%s %s", prefix, suffix)
}

// UploadActivity uploads the activity and sets the gear ID and commute status for an activity
func UploadActivity(accessToken string, activityTime time.Time, activityFilename string, gearID string, isCommute bool, isWalk bool, isRun bool) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	f, err := os.Open(activityFilename)
	if err != nil {
		return fmt.Errorf("failed to open %q: %v", activityFilename, err)
	}
	activityName := GetActivityNameForTime(activityTime, isWalk, isRun)
	opts := strava.CreateUploadOpts{
		Name:     optional.NewString(activityName),
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
	var activityID int64
	for {
		if err != nil {
			var msg string
			if resp != nil {
				body, _ := ioutil.ReadAll(resp.Body)
				msg = string(body)
			}
			return fmt.Errorf("%v %s", err, msg)
		}
		if upload.ActivityId != 0 {
			logger.GetLogger().Printf("Got activity Id: %d", upload.ActivityId)
			activityID = upload.ActivityId
			break
		}
		duplicateID, uploadError := handleError(upload.Error_)
		if uploadError != nil {
			return uploadError
		}
		if duplicateID != 0 {
			activityID = duplicateID
			break
		}
		time.Sleep(1 * time.Second)
		logger.GetLogger().Printf(" checking on status of upload Id %d...", upload.Id)
		upload, resp, err = client.UploadsApi.GetUploadById(ctx, upload.Id)
	}
	err = UpdateActivity(accessToken, activityID, activityName, gearID, isCommute, isWalk, isRun)
	if err != nil {
		return err
	}
	fmt.Printf("Uploaded and/or updated: https://www.strava.com/activities/%d\n", activityID)
	return nil
}

// UpdateActivity sets various parameters on an activity that already exists
func UpdateActivity(accessToken string, activityID int64, activityName string, gearID string, isCommute bool, isWalk bool, isRun bool) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)
	update := strava.UpdatableActivity{
		GearId: gearID,
		Name:   activityName,
	}
	if isWalk {
		activityType := strava.WALK_ActivityType
		update.Type_ = &activityType
	}
	if isRun {
		activityType := strava.RUN_ActivityType
		update.Type_ = &activityType
	}
	if isCommute {
		update.Commute = true
	}

	opts := strava.UpdateActivityByIdOpts{
		Body: optional.NewInterface(update),
	}
	logger.GetLogger().Printf("Setting activity with id %d to gearID: %s, commute: %t, walk: %t", activityID, gearID, isCommute, isWalk)
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

func handleError(uploadError string) (int64, error) {
	if uploadError == "" {
		return 0, nil
	}

	switch {
	case strings.Contains(uploadError, "duplicate"):
		logger.GetLogger().Printf("Skipping duplicate upload: %s", uploadError)
		re := regexp.MustCompile(`.+? duplicate of(?: an uploading)? activity \(?(\d+)\)?`)
		match := re.FindStringSubmatch(uploadError)
		if match == nil {
			return 0, fmt.Errorf("Failed to match activity ID in string: %s", uploadError)
		}
		activityID, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Failed to parse int from match in string: %s, with: %s", uploadError, err)
		}
		return activityID, nil
	default:
		return 0, fmt.Errorf("Strava upload failed: %s", uploadError)
	}
}
