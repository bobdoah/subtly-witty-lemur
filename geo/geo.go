package geo

import (
	"math"
	"strconv"
	"time"

	"github.com/philhofer/tcx"

	"github.com/bobdoah/subtly-witty-lemur/logger"
)

const (

	// EarthRadius according to wikipedia
	EarthRadius = 6371
)

// Point on the Earth
type Point struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Deg2rad converts a Degree angle to radians
func Deg2rad(deg float64) float64 {
	return float64(math.Pi * deg / 180)
}

// Distance calculates distance between two points
func Distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radlat1 := Deg2rad(lat1)
	radlat2 := Deg2rad(lat2)

	dLon := Deg2rad(float64(lon1 - lon2))
	dLat := Deg2rad(float64(lat1 - lat2))

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(radlat1)*math.Cos(radlat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

//StartIsCloseToOneOf compares the start of TCX to a list of points
func StartIsCloseToOneOf(tcxdb *tcx.TCXDB, points map[string]Point) bool {
	activity := tcxdb.Acts.Act[0]
	trackpoint := activity.Laps[0].Trk.Pt[0]
	logger.DebugLogger.Printf("Activity %s, first trackpoint Latitude %f, Longitude %f\n", activity.Id.Format(time.RFC3339), trackpoint.Lat, trackpoint.Long)
	return TrackpointIsCloseToOneOf(trackpoint, points)
}

//EndIsCloseToOneOf compares the end of TCX to a list of points
func EndIsCloseToOneOf(tcxdb *tcx.TCXDB, points map[string]Point) bool {
	activity := tcxdb.Acts.Act[0]
	laps := activity.Laps
	lastLapTrackpoints := laps[len(laps)-1].Trk.Pt
	lastTrackpoint := lastLapTrackpoints[len(lastLapTrackpoints)-1]
	logger.DebugLogger.Printf("Activity %s, last trackpoint Latitude %f, Longitude %f\n", activity.Id.Format(time.RFC3339), lastTrackpoint.Lat, lastTrackpoint.Long)
	return TrackpointIsCloseToOneOf(lastTrackpoint, points)

}

//TrackpointIsCloseToOneOf compares a TCX trackpoint to a list of latitude and longitude points and return true if it is close to any of them
func TrackpointIsCloseToOneOf(trackpoint tcx.Trackpoint, points map[string]Point) bool {
	var closeTo float64 = 5
	for postcode, point := range points {
		dist := Distance(trackpoint.Lat, trackpoint.Long, point.Latitude, point.Longitude)
		if dist < closeTo {
			logger.DebugLogger.Printf("Trackpoint is less than %skm from %s\n", strconv.FormatFloat(closeTo, 'f', -1, 64), postcode)
			return true
		}
	}
	return false
}
