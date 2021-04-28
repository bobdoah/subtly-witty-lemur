package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bobdoah/subtly-witty-lemur/geo"
	"github.com/bobdoah/subtly-witty-lemur/strava"
	"github.com/philhofer/tcx"
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
		fmt.Printf("id: %s sport: %s\n", act.Id.Format(time.RFC3339), act.Sport)
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
		fmt.Printf("Postcode: %s is at lat: %f, lon: %f\n", postcode, point.Latitude, point.Longitude)
	}
	return nil
}

func main() {
	var clientID, clientSecret string

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
				Name:        "client-id",
				Usage:       "Client ID for strava api",
				Required:    true,
				Destination: &clientId,
			},
			&cli.StringFlag{
				Name:        "client-secret",
				Usage:       "Client secret for strava api",
				Required:    true,
				Destination: &clientSecret,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() > 0 {
				var i int
				homePoints, err := geo.GetPointsFromPostcodes(c.StringSlice("home"))
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
					activitySummaries, err := strava.GetActivityForTime(accessToken, activity.Id)
					if err != nil {
						return err
					}
					for _, activitySummary := range activitySummaries {
						fmt.Printf("Existing Strava activity, id: %d, name: %s\n", activitySummary.Id, activitySummary.Name)
					}
				}
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
