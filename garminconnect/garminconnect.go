package garminconnect

import (
	"fmt"
	"time"

	"github.com/abrander/garmin-connect"
	"github.com/bobdoah/subtly-witty-lemur/state"
)

// GetCalenderItemForTime returns a Calendar Item for a given start time
func GetCalenderItemForTime(startTime time.Time) (*connect.CalendarItem, error) {

	afterTime := startTime.Add(time.Minute * -10)
	beforeTime := startTime.Add(time.Minute * 10)
	fmt.Printf("Checking Garmin Connect for activities between %v and %v\n", beforeTime, afterTime)

	year, month, day := startTime.Date()
	calendar, err := state.AuthState.Garmin.CalendarWeek(int(year), int(month), int(day))
	if err != nil {
		return nil, err
	}
	for _, calendarItem := range calendar.CalendarItems {

	}
	return nil, nil
}
