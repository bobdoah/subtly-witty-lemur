package geo

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type result struct {
	Status int   `json:"status"`
	Result Point `json:"result"`
}

// GetPointFromPostcode returns Latitude and Longitude point for a given postcode
func GetPointFromPostcode(postcode string) (Point, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result{}).
		Get("https://api.postcodes.io/postcodes/" + postcode)
	if err != nil {
		return Point{}, err
	}
	response := resp.Result().(*result)

	return response.Result, nil
}

// GetPointsFromPostcodes returns a map of postcodes to Latitude and Longitude points
func GetPointsFromPostcodes(postcodes []string) (map[string]Point, error) {
	points := make(map[string]Point)
	for _, postcode := range postcodes {
		point, err := GetPointFromPostcode(postcode)
		fmt.Printf("Postcode %s, point is Latitude: %f, Longitude %f\n", postcode, point.Latitude, point.Longitude)
		if err != nil {
			return points, err
		}
		points[postcode] = point
	}
	return points, nil
}
