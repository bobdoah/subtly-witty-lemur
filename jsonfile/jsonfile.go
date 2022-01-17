package jsonfile

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/buger/jsonparser"

	"github.com/bobdoah/subtly-witty-lemur/logger"
)

// GetSportFromJSONFile parses an Endomondo JSON file and returns the sport
func GetSportFromJSONFile(filename *string) (string, error) {
	jsonfile, err := os.Open(*filename)
	defer jsonfile.Close()
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(jsonfile)
	if err != nil {
		return "", err
	}
	var sport string
	jsonparser.ArrayEach(bytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		sportValue, err := jsonparser.GetString(value, "sport")
		logger.GetLogger().Printf("sport: %s", sportValue)
		if sportValue != "" {
			sport = sportValue
		}
	})
	if sport == "" {
		err = fmt.Errorf("Sport not found")
	}
	return sport, err
}
