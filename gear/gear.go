package gear

// Collection holds the details for the gear
type Collection struct {
	CommuteBike  gear
	MountainBike gear
	RoadBike     gear
}

// Gear holds the strings for the names of the gear
type gear struct {
	Name       string
	GarminUUID string
	StravaID   string
}
