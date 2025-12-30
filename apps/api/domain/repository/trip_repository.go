package repository

import "fukuoka-ai-api/domain/entity"

type TripRepository interface {
	EnsureUser(userID string) error
	CreateTrip(trip *entity.Trip) error
	GetTrip(tripID string) (*entity.Trip, error)
	GetTripByShareID(shareID string) (*entity.Trip, error)
	CreateTripPlace(place *entity.TripPlace) error
	GetTripPlaces(tripID string) ([]*entity.TripPlace, error)
	GetTripPlacesByIDs(tripID string, placeIDs []string) ([]*entity.TripPlace, error)
	UpdateTripPlace(place *entity.TripPlace) error
	CreateShare(shareID, tripID string) error
}

