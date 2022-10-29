package storage

import "database/sql"

type StorageI interface {
	Min(days int) (*sql.Rows, error)
	Max(days int) (*sql.Rows, error)
	Average(days int) (*sql.Rows, error)
}

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Min(days int) (*sql.Rows, error) {
	return s.db.Query("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY main_temp ASC limit 1", days)
}

func (s *Storage) Max(days int) (*sql.Rows, error) {
	return s.db.Query("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY main_temp DESC limit 1", days)
}

func (s *Storage) Average(days int) (*sql.Rows, error) {
	return s.db.Query("select ROUND(AVG(main_temp), 2), city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY)", days)
}
