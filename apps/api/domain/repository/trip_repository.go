package repository

import "fukuoka-ai-api/domain/model"

type TripRepository interface {
	EnsureUser(userID string) error
	CreateTrip(trip *model.Trip) error
	GetTrip(tripID string) (*model.Trip, error)
	GetTripByShareID(shareID string) (*model.Trip, error)
	CreateTripPlace(place *model.TripPlace) error
	GetTripPlaces(tripID string) ([]*model.TripPlace, error)
	GetTripPlacesByIDs(tripID string, placeIDs []string) ([]*model.TripPlace, error)
	UpdateTripPlace(place *model.TripPlace) error
	CreateShare(shareID, tripID string) error
}

