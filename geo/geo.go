package geo

import (
	"math"
)

const (

	// EarthRadius according to wikipedia
	EarthRadius = 6371
	// DistanceFromHome in km to be considered starting from home
	DistanceFromHome = 5
	// DistanceFromWork in km to be considered starting from work
	DistanceFromWork = 5
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
