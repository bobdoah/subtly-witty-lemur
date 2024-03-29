package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/philhofer/tcx"
	"github.com/urfave/cli/v2"

	"github.com/bobdoah/subtly-witty-lemur/garminconnect"
	"github.com/bobdoah/subtly-witty-lemur/gear"
	"github.com/bobdoah/subtly-witty-lemur/geo"
	"github.com/bobdoah/subtly-witty-lemur/jsonfile"
	"github.com/bobdoah/subtly-witty-lemur/logger"
	"github.com/bobdoah/subtly-witty-lemur/state"
	"github.com/bobdoah/subtly-witty-lemur/stravautils"
)

func readTcx(filepath string) (*tcx.TCXDB, error) {
	db, err := tcx.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	nacts := len(db.Acts.Act)
	if nacts > 0 {
		act := db.Acts.Act[0]
		logger.GetLogger().Printf("id: %s sport: %s\n", act.Id.Format(time.RFC3339), act.Sport)
	}

	return db, nil
}

func isCommute(db *tcx.TCXDB, homePoints map[string]geo.Point, workPoints map[string]geo.Point) bool {
	return ((geo.StartIsCloseToOneOf(db, homePoints) && geo.EndIsCloseToOneOf(db, workPoints)) ||
		(geo.StartIsCloseToOneOf(db, workPoints) && geo.EndIsCloseToOneOf(db, homePoints))) && !isWeekendRide(db)
}

func isWeekendRide(db *tcx.TCXDB) bool {
	activity := db.Acts.Act[0]
	trackpoint := activity.Laps[0].Trk.Pt[0]
	weekday := trackpoint.Time.Weekday()
	logger.GetLogger().Printf("Weekday: %s", weekday)
	switch weekday {
	case
		time.Sunday,
		time.Saturday:
		return true
	}
	return false
}

func printLatLons(postcodes []string) error {
	for _, postcode := range postcodes {
		point, err := geo.GetPointFromPostcode(postcode)
		if err != nil {
			return err
		}
		logger.GetLogger().Printf("Postcode: %s is at lat: %f, lon: %f\n", postcode, point.Latitude, point.Longitude)
	}
	return nil
}

func getJsonfileName(tcxFilename string, directoryName *string) string {
	jsonFilename := tcxFilename[:len(tcxFilename)-len(filepath.Ext(tcxFilename))] + ".json"
	if directoryName != nil {
		jsonFilename = filepath.Join(*directoryName, filepath.Base(jsonFilename))
	}
	logger.GetLogger().Printf("tcx: %s, json: %s", tcxFilename, jsonFilename)
	return jsonFilename
}

func processUploadFile(filename string, homePoints map[string]geo.Point, workPoints map[string]geo.Point, gear gear.Collection, uploadGarmin bool, uploadStrava bool, jsonfileDirectory *string, uploadWalk bool, uploadRun bool) error {
	db, err := readTcx(filename)
	activity := db.Acts.Act[0]
	var garminGearUUID string
	var stravaGearID string
	var rideIsCommute bool
	var isWalk bool = false
	var isRun bool = false
	jsonFilename := getJsonfileName(filename, jsonfileDirectory)
	var jsonFileSport string
	jsonFileSport, err = jsonfile.GetSportFromJSONFile(&jsonFilename)
	if err != nil {
		return err
	}
	if jsonFileSport == "WALKING" {
		logger.GetLogger().Printf("id: %s JSON sport: %s\n", activity.Id.Format(time.RFC3339), jsonFileSport)
		isWalk = true
		if !uploadWalk {
			logger.GetLogger().Printf("skipping upload for walking activity\n")
			return nil
		}
	} else if jsonFileSport == "RUNNING" {
		logger.GetLogger().Printf("id: %s JSON sport: %s\n", activity.Id.Format(time.RFC3339), jsonFileSport)
		isRun = true
		if !uploadRun {
			logger.GetLogger().Printf("skipping upload for running activity\n")
			return nil
		}
	} else {
		garminGearUUID = gear.RoadBike.GarminUUID
		stravaGearID = gear.RoadBike.StravaID
		if err != nil {
			return err
		}
		rideIsCommute = jsonFileSport == "CYCLING_TRANSPORTATION" || isCommute(db, homePoints, workPoints)
		var not string = ""
		if rideIsCommute {
			garminGearUUID = gear.CommuteBike.GarminUUID
			stravaGearID = gear.CommuteBike.StravaID
		} else {
			not = "not "
		}
		rideIsMTB := jsonFileSport == "MOUNTAIN_BIKING"
		var notMTB string = ""
		if rideIsMTB {
			garminGearUUID = gear.MountainBike.GarminUUID
			stravaGearID = gear.MountainBike.StravaID
		} else {
			notMTB = "not "
		}
		fmt.Printf("id: %s sport: %s is %sa commute, is %smtb\n", activity.Id.Format(time.RFC3339), activity.Sport, not, notMTB)
	}
	startTime := activity.Laps[0].Trk.Pt[0].Time
	activitySummaries, err := stravautils.GetActivityForTime(state.AuthState.StravaAccessToken, startTime)
	if err != nil {
		return err
	}
	for _, activitySummary := range activitySummaries {
		fmt.Printf("Existing Strava activity, id: %d, name: %s\n", activitySummary.Id, activitySummary.Name)
	}
	calendarItem, err := garminconnect.GetCalenderItemForTime(startTime)
	if err != nil {
		return err
	}
	if calendarItem != nil {
		fmt.Printf("Existing Garmin activity, id: %v, title: %s\n", calendarItem.ID, calendarItem.Title)
	}
	if uploadGarmin && calendarItem == nil {
		id, err := garminconnect.ActivityUpload(filename)
		if err != nil {
			return err
		}
		if garminGearUUID != "" {
			err = state.AuthState.Garmin.GearLink(garminGearUUID, id)
			if err != nil {
				fmt.Printf("Failed to set Gear for activity %d", id)
				return err
			}
			fmt.Printf("Successfully set Gear for activity %d", id)
			if err != nil {
				return err
			}
		} else {
			logger.GetLogger().Printf("Not setting Garmin gear for activity %d", id)
		}
	} else if calendarItem != nil {
		logger.GetLogger().Printf("Existing Garmin activity found. Not uploading\n")

	} else {
		fmt.Printf("Would upload to garmin %s with gear ID %s, but skipping\n", filename, garminGearUUID)
	}
	if uploadStrava && len(activitySummaries) == 0 {
		err = stravautils.UploadActivity(state.AuthState.StravaAccessToken, activity.Id, filename, stravaGearID, rideIsCommute, isWalk, isRun)
		if err != nil {
			return err
		}
	} else if len(activitySummaries) > 0 {
		logger.GetLogger().Printf("Existing Strava activity found. Not uploading\n")
	} else {
		fmt.Printf("Would upload to strava %s with gear ID %s, but skipping\n", filename, stravaGearID)
	}

	return nil
}

func main() {
	gear := gear.Collection{}
	var jsonfileDirectory string
	app := &cli.App{
		Name:  "upload-tcx",
		Usage: "upload a tcx file to somewhere",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "home",
				Usage: "Postcodes of home starting points",
			},
			&cli.StringSliceFlag{
				Name:  "work",
				Usage: "Postcodes of work starting points",
			},
			&cli.StringFlag{
				Name:        "mtb",
				Usage:       "name of the mountain bike in Garmin and Strava",
				Destination: &gear.MountainBike.Name,
			},
			&cli.StringFlag{
				Name:        "commute",
				Usage:       "name of the commute bike in Garmin and Strava",
				Destination: &gear.CommuteBike.Name,
			},
			&cli.StringFlag{
				Name:        "road",
				Usage:       "name of the road bike in Garmin and Strava",
				Destination: &gear.RoadBike.Name,
			},
			&cli.StringFlag{
				Name:        "state-file",
				Usage:       "State file for Strava API",
				Destination: &state.StateFile,
				Value:       state.Filename(),
			},
			&cli.BoolFlag{
				Name:        "debug",
				Value:       false,
				Destination: &logger.Enabled,
			},
			&cli.BoolFlag{
				Name:  "no-garmin",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "no-strava",
				Value: false,
			},
			&cli.StringFlag{
				Name:        "json-files-dir",
				Usage:       "directory for JSON files (if not cwd",
				Destination: &jsonfileDirectory,
			},
			&cli.BoolFlag{
				Name:  "skip-walking",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "skip-running",
				Value: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "authenticate-with-strava",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "client-id",
						Usage:    "Client ID for strava api",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "client-secret",
						Usage:    "Client secret for strava api",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "port",
						Usage: "port to bind local http server to",
						Value: 8080,
					},
				},
				Action: stravautils.Authenticate,
			},
			{
				Name: "authenticate-with-garmin-connect",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "email",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Required: true,
					},
				},
				Action: garminconnect.Authenticate,
			},
			{
				Name:   "signout-of-garmin-connect",
				Action: garminconnect.Signout,
			},
			{
				Name: "summary",
				Action: func(c *cli.Context) error {
					if c.NArg() > 0 {
						var i int
						homePoints, err := geo.GetPointsFromPostcodes(c.StringSlice("home"))
						if err != nil {
							return err
						}
						workPoints, err := geo.GetPointsFromPostcodes(c.StringSlice("work"))
						if err != nil {
							return err
						}
						for i = 0; i < c.Args().Len(); i++ {
							db, err := readTcx(c.Args().Get(i))
							activity := db.Acts.Act[0]
							if err != nil {
								return err
							}
							rideIsCommute := isCommute(db, homePoints, workPoints)
							var not string = ""
							if !rideIsCommute {
								not = "not "
							}
							fmt.Printf("id: %s sport: %s is %sa commute\n", activity.Id.Format(time.RFC3339), activity.Sport, not)
						}
					}
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() > 0 {
				state.LoadState()
				err := garminconnect.GetGearUUIDs(&gear)
				if err != nil {
					return err
				}
				err = stravautils.GetGearIds(state.AuthState.StravaAccessToken, &gear)
				if err != nil {
					return err
				}
				homePoints, err := geo.GetPointsFromPostcodes(c.StringSlice("home"))
				if err != nil {
					return err
				}
				workPoints, err := geo.GetPointsFromPostcodes(c.StringSlice("work"))
				if err != nil {
					return err
				}
				uploadGarmin := !(c.Bool("no-garmin"))
				uploadStrava := !(c.Bool("no-strava"))
				uploadWalk := !(c.Bool("skip-walking"))
				uploadRun := !(c.Bool("skip-running"))
				uploadFiles := c.Args().Slice()
				for _, uploadFile := range uploadFiles {
					err = processUploadFile(uploadFile, homePoints, workPoints, gear, uploadGarmin, uploadStrava, &jsonfileDirectory, uploadWalk, uploadRun)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}
