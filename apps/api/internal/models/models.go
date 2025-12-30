package models

type User struct {
	ID        string
	CreatedAt string
}

type Trip struct {
	ID        string
	UserID    string
	Title     string
	StartTime string
	CreatedAt string
}

type TripPlace struct {
	ID           string
	TripID       string
	PlaceID      string
	Name         string
	Lat          float64
	Lng          float64
	Kind         string // must, recommended, start
	StayMinutes int
	OrderIndex   int
	Reason       string
	ReviewSummary string
	PhotoURL     string
}

type Share struct {
	ShareID   string
	TripID    string
	CreatedAt string
}

type Place struct {
	ID            string
	PlaceID       string
	Name          string
	Lat           float64
	Lng           float64
	Kind          string
	StayMinutes   int
	OrderIndex    int
	TimeRange     string
	Reason        string
	ReviewSummary string
	PhotoURL      string
	Category      string
}

type Route struct {
	Polyline string
}


