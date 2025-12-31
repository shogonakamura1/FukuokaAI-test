package repositories

import (
	"database/sql"

	"fukuoka-ai-api/models"
)

type ITripRepository interface {
	EnsureUser(userID string) error
	CreateTrip(trip *models.Trip) error
	GetTrip(tripID string) (*models.Trip, error)
	GetTripByShareID(shareID string) (*models.Trip, error)
	CreateTripPlace(place *models.TripPlace) error
	GetTripPlaces(tripID string) ([]*models.TripPlace, error)
	GetTripPlacesByIDs(tripID string, placeIDs []string) ([]*models.TripPlace, error)
	UpdateTripPlace(place *models.TripPlace) error
	CreateShare(shareID, tripID string) error
}

type TripRepository struct {
	db *sql.DB
}

func NewTripRepository(db *sql.DB) ITripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) EnsureUser(userID string) error {
	_, err := r.db.Exec(
		"INSERT OR IGNORE INTO users (id) VALUES (?)",
		userID,
	)
	return err
}

func (r *TripRepository) CreateTrip(trip *models.Trip) error {
	_, err := r.db.Exec(
		"INSERT INTO trips (id, user_id, title, start_time) VALUES (?, ?, ?, ?)",
		trip.ID, trip.UserID, trip.Title, trip.StartTime,
	)
	return err
}

func (r *TripRepository) GetTrip(tripID string) (*models.Trip, error) {
	trip := &models.Trip{}
	err := r.db.QueryRow(
		"SELECT id, user_id, title, start_time, created_at FROM trips WHERE id = ?",
		tripID,
	).Scan(&trip.ID, &trip.UserID, &trip.Title, &trip.StartTime, &trip.CreatedAt)
	if err != nil {
		return nil, err
	}
	return trip, nil
}

func (r *TripRepository) GetTripByShareID(shareID string) (*models.Trip, error) {
	trip := &models.Trip{}
	err := r.db.QueryRow(
		`SELECT t.id, t.user_id, t.title, t.start_time, t.created_at 
		 FROM trips t 
		 JOIN shares s ON t.id = s.trip_id 
		 WHERE s.share_id = ?`,
		shareID,
	).Scan(&trip.ID, &trip.UserID, &trip.Title, &trip.StartTime, &trip.CreatedAt)
	if err != nil {
		return nil, err
	}
	return trip, nil
}

func (r *TripRepository) CreateTripPlace(place *models.TripPlace) error {
	_, err := r.db.Exec(
		`INSERT INTO trip_places 
		 (id, trip_id, place_id, name, lat, lng, kind, stay_minutes, order_index, reason, review_summary, photo_url) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		place.ID, place.TripID, place.PlaceID, place.Name, place.Lat, place.Lng,
		place.Kind, place.StayMinutes, place.OrderIndex, place.Reason, place.ReviewSummary, place.PhotoURL,
	)
	return err
}

func (r *TripRepository) GetTripPlaces(tripID string) ([]*models.TripPlace, error) {
	rows, err := r.db.Query(
		`SELECT id, trip_id, place_id, name, lat, lng, kind, stay_minutes, order_index, reason, review_summary, photo_url 
		 FROM trip_places 
		 WHERE trip_id = ? 
		 ORDER BY order_index`,
		tripID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := []*models.TripPlace{}
	for rows.Next() {
		place := &models.TripPlace{}
		err := rows.Scan(
			&place.ID, &place.TripID, &place.PlaceID, &place.Name,
			&place.Lat, &place.Lng, &place.Kind, &place.StayMinutes,
			&place.OrderIndex, &place.Reason, &place.ReviewSummary, &place.PhotoURL,
		)
		if err != nil {
			return nil, err
		}
		places = append(places, place)
	}
	return places, nil
}

func (r *TripRepository) GetTripPlacesByIDs(tripID string, placeIDs []string) ([]*models.TripPlace, error) {
	if len(placeIDs) == 0 {
		return []*models.TripPlace{}, nil
	}

	query := `SELECT id, trip_id, place_id, name, lat, lng, kind, stay_minutes, order_index, reason, review_summary, photo_url 
			  FROM trip_places 
			  WHERE trip_id = ? AND id IN (`
	args := []interface{}{tripID}
	for i, id := range placeIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ") ORDER BY order_index"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := []*models.TripPlace{}
	placeMap := make(map[string]*models.TripPlace)
	for rows.Next() {
		place := &models.TripPlace{}
		err := rows.Scan(
			&place.ID, &place.TripID, &place.PlaceID, &place.Name,
			&place.Lat, &place.Lng, &place.Kind, &place.StayMinutes,
			&place.OrderIndex, &place.Reason, &place.ReviewSummary, &place.PhotoURL,
		)
		if err != nil {
			return nil, err
		}
		placeMap[place.ID] = place
	}

	// 指定された順序で返す
	for _, id := range placeIDs {
		if place, ok := placeMap[id]; ok {
			places = append(places, place)
		}
	}

	return places, nil
}

func (r *TripRepository) UpdateTripPlace(place *models.TripPlace) error {
	_, err := r.db.Exec(
		`UPDATE trip_places 
		 SET stay_minutes = ?, order_index = ? 
		 WHERE id = ?`,
		place.StayMinutes, place.OrderIndex, place.ID,
	)
	return err
}

func (r *TripRepository) CreateShare(shareID, tripID string) error {
	_, err := r.db.Exec(
		"INSERT INTO shares (share_id, trip_id) VALUES (?, ?)",
		shareID, tripID,
	)
	return err
}

