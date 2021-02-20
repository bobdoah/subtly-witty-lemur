package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bobdoah/subtly-witty-lemur/geo"
	"github.com/philhofer/tcx"
	"github.com/urfave/cli/v2"
)

func readTcx(filepath string) error {
	db, err := tcx.ReadFile(filepath)
	if err != nil {
		return err
	}
	nacts := len(db.Acts.Act)
	if nacts > 0 {
		act := db.Acts.Act[0]
		fmt.Printf("id: %s sport: %s\n", act.Id.Format(time.RFC3339), act.Sport)
	}

	return nil
}

func printLatLons(postcodes []string) error {
	for _, postcode := range postcodes {
		coords := geo.Point{}
		err := geo.GetPointFromPostcode(postcode, &coords)
		if err != nil {
			return err
		}
		fmt.Printf("Postcode: %s is at lat: %f, lon: %f\n", postcode, coords.Latitude, coords.Longitude)
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
		},
		Action: func(c *cli.Context) error {
			if c.NArg() > 0 {
				var i int
				for i = 0; i < c.Args().Len(); i++ {
					readTcx(c.Args().Get(i))
				}
			}
			return printLatLons(c.StringSlice("home"))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
