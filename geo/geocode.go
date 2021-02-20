package geo

import (
	"github.com/go-resty/resty/v2"
)

type result struct {
	Status int   `json:"status"`
	Result Point `json:"result"`
}

// GetPointFromPostcode returns Latitude and Longitude point for a given postcode
func GetPointFromPostcode(postcode string, coords *Point) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&result{}).
		Get("https://api.postcodes.io/postcodes/" + postcode)
	if err != nil {
		return err
	}
	response := resp.Result().(*result)
	*coords = response.Result
	return nil
}
