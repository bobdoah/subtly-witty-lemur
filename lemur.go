package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/philhofer/tcx"
	"github.com/urfave/cli/v2"
)

const (

	// EarthRadius according to wikipedia
	EarthRadius = 6371
	// DistanceFromHome in km to be considered starting from home
	DistanceFromHome = 5
	// DistanceFromWork in km to be considered starting from work
	DistanceFromWork = 5
)

type point struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type result struct {
	Status int   `json:"status"`
	Point  point `json:"result"`
}

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

func getLatLongFromPostcode(postcode string, coords *point) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result{}).
		Get("https://api.postcodes.io/postcodes/" + postcode)
	if err != nil {
		return err
	}
	response := resp.Result().(*result)
	*coords = response.Point
	return nil
}

func printLatLons(postcodes []string) error {
	for _, postcode := range postcodes {
		coords := point{}
		err := getLatLongFromPostcode(postcode, &coords)
		if err != nil {
			return err
		}
		fmt.Printf("Postcode: %s is at lat: %f, lon: %f\n", postcode, coords.Latitude, coords.Longitude)
	}
	return nil
}

func deg2rad(deg float64) float64 {
	return float64(math.Pi * deg / 180)
}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radlat1 := deg2rad(lat1)
	radlat2 := deg2rad(lat2)

	dLon := deg2rad(float64(lon1 - lon2))
	dLat := deg2rad(float64(lat1 - lat2))

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(radlat1)*math.Cos(radlat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
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
