package jsonfile

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/bobdoah/subtly-witty-lemur/logger"
)

// Workout represents a workout in the JSON file
type Workout struct {
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Source      string    `json:"source"`
	CreatedDate time.Time `json:"created_date"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    float64   `json:"duaration_s"`
	Distance    float64   `json:"distance_km"`
	Calories    float64   `json:"calories_kcal"`
	AltitudeMin float64   `json:"altitude_min_m"`
	AltitudeMax float64   `json:"altitude_max_m"`
	Speed       float64   `json:"speed_avg_kmh"`
	Ascend      float64   `json:"ascend_m"`
	Descend     float64   `json:"descend_m"`
	Points      [][]Point `json:"points"`
}

// Point represents a geographic point recorded during the workout
type Point struct {
	Location  [][]Location `json:"location,omitempty"`
	Altitude  float64      `json:"altitude,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

// Location represents a geographic location in latitude and longitude
type Location struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// ReadJsonfile parses an Endomondo JSON file and return workouts
func ReadJsonfile(filename *string) (*[]Workout, error) {
	jsonfile, err := os.Open(*filename)
	defer jsonfile.Close()
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(jsonfile)
	if err != nil {
		return nil, err
	}
	var workouts []Workout
	json.Unmarshal(bytes, &workouts)
	for _, workout := range workouts {
		logger.GetLogger().Printf("name: %s sport: %s\n", workout.Name, workout.Sport)
	}
	return &workouts, nil
}
