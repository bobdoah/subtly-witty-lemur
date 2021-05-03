package garminconnect

import (
	"fmt"
	"strings"
	"time"

	"github.com/abrander/garmin-connect"
	"github.com/bobdoah/subtly-witty-lemur/logger"
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
		startTime := calendarItem.StartTimestampLocal
		logger.GetLogger().Printf("Activity %v, start time %v\n", calendarItem.ID, startTime)
		if startTime.Before(beforeTime) && startTime.After(afterTime) {
			return &calendarItem, nil
		}
	}
	return nil, nil
}

// getGearUUID returns the Garmin Connect ID for a given gear name string
func getGearUUID(gearName string) (*string, error) {
	gear, err := state.AuthState.Garmin.Gear(0)
	if err != nil {
		return nil, err
	}
	for _, g := range gear {
		logger.GetLogger().Printf("Checking %s matches %s", g.DisplayName, gearName)
		if strings.EqualFold(g.DisplayName, gearName) {
			logger.GetLogger().Printf("Matched Type %s ID %d", g.DisplayName, g.Uuid)
			return &g.Uuid, nil
		}
	}
	return nil, nil
}

// GetGearUUIDs gets the UUIDs for the gear names supplied
func GetGearUUIDs(gear *GearList) error {
	commuteUUID, err := getGearUUID(gear.CommuteBike.Name)
	if err != nil {
		return err
	}
	gear.CommuteBike.GarminUUID = commuteUUID
	roadUUID, err := getGearUUID(gear.RoadBike.Name)
	if err != nil {
		return err
	}
	gear.RoadBike.GarminUUID = roadUUID
	mountainUUID, err := getGearUUID(gear.MountainBike.Name)
	if err != nil {
		return err
	}
	gear.MountainBike.GarminUUID = mountainUUID
	return nil
}
