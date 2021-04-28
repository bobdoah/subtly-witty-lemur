package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bobdoah/subtly-witty-lemur/garminconnect"
	"github.com/bobdoah/subtly-witty-lemur/geo"
	"github.com/bobdoah/subtly-witty-lemur/logger"
	"github.com/bobdoah/subtly-witty-lemur/state"
	"github.com/bobdoah/subtly-witty-lemur/stravautils"
	"github.com/philhofer/tcx"
	"github.com/strava/go.strava"
	"github.com/urfave/cli/v2"
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
		(geo.StartIsCloseToOneOf(db, workPoints) && geo.EndIsCloseToOneOf(db, homePoints)))
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

func main() {
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
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name: "authenticate-with-strava",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "client-id",
						Usage:       "Client ID for strava api",
						Destination: &strava.ClientId,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "client-secret",
						Usage:       "Client secret for strava api",
						Destination: &strava.ClientSecret,
						Required:    true,
					},
					&cli.IntFlag{
						Name:  "port",
						Usage: "port to bind local http server to",
						Value: 8080,
					},
				},
				Action: stravautils.Authenticate,
			},
			&cli.Command{
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
			&cli.Command{
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
							if isCommute(db, homePoints, workPoints) {
								fmt.Printf("id: %s sport: %s is a commute\n", activity.Id.Format(time.RFC3339), activity.Sport)
							}
						}
					}
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() > 0 {
				var i int
				state.LoadState()
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
					if isCommute(db, homePoints, workPoints) {
						fmt.Printf("id: %s sport: %s is a commute\n", activity.Id.Format(time.RFC3339), activity.Sport)
					}
					startTime := activity.Laps[0].Trk.Pt[0].Time
					activitySummaries, err := stravautils.GetActivityForTime(state.AuthState.Strava.AccessToken, startTime)
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
