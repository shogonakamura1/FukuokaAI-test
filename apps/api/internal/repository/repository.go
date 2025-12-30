package repository

import (
	"database/sql"
	"fukuoka-ai-api/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EnsureUser(userID string) error {
	_, err := r.db.Exec(
		"INSERT OR IGNORE INTO users (id) VALUES (?)",
		userID,
	)
	return err
}

func (r *Repository) CreateTrip(trip *models.Trip) error {
	_, err := r.db.Exec(
		"INSERT INTO trips (id, user_id, title, start_time) VALUES (?, ?, ?, ?)",
		trip.ID, trip.UserID, trip.Title, trip.StartTime,
	)
	return err
}

func (r *Repository) GetTrip(tripID string) (*models.Trip, error) {
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

func (r *Repository) GetTripByShareID(shareID string) (*models.Trip, error) {
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

func (r *Repository) CreateTripPlace(place *models.TripPlace) error {
	_, err := r.db.Exec(
		`INSERT INTO trip_places 
		 (id, trip_id, place_id, name, lat, lng, kind, stay_minutes, order_index, reason, review_summary, photo_url) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		place.ID, place.TripID, place.PlaceID, place.Name, place.Lat, place.Lng,
		place.Kind, place.StayMinutes, place.OrderIndex, place.Reason, place.ReviewSummary, place.PhotoURL,
	)
	return err
}

func (r *Repository) GetTripPlaces(tripID string) ([]*models.TripPlace, error) {
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

func (r *Repository) GetTripPlacesByIDs(tripID string, placeIDs []string) ([]*models.TripPlace, error) {
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

func (r *Repository) UpdateTripPlace(place *models.TripPlace) error {
	_, err := r.db.Exec(
		`UPDATE trip_places 
		 SET stay_minutes = ?, order_index = ? 
		 WHERE id = ?`,
		place.StayMinutes, place.OrderIndex, place.ID,
	)
	return err
}

func (r *Repository) CreateShare(shareID, tripID string) error {
	_, err := r.db.Exec(
		"INSERT INTO shares (share_id, trip_id) VALUES (?, ?)",
		shareID, tripID,
	)
	return err
}


