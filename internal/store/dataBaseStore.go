package store

import (
	"database/sql"
	"github.com/rs/zerolog/log"
)

// DB a global variable to hold the database connection
var DB *sql.DB

// BatchValues a struct to hold the values for batch insert
type BatchValues struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UUID        string `json:"uuid"`
}

// SQL statement to create the table
const createTableSQL = `
		CREATE TABLE IF NOT EXISTS urls (
			uuid TEXT,
			short_url TEXT NOT NULL,
			original_url TEXT NOT NULL
		);`

// SQL statement to insert into the table
const insertSQL = `INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`

// SQL statement to delete from the table
const deleteSQL = `DELETE FROM urls WHERE short_url = $1;`

// SQL statement to select from the table
const selectSQL = `SELECT original_url FROM urls WHERE short_url = $1;`

// NewDBStore creates a new store
func InitDB(db *sql.DB) error {
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	log.Info().Msg("Table 'urls' created successfully")
	return nil
}

// SaveToDB saves a URL to the database
func SaveToDB(shortURL, originalURL, uuid string) {
	if uuid == "" {
		uuid = GenerateUUID()
	}
	_, err := DB.Exec(insertSQL, uuid, shortURL, originalURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save URL")
	}
}

// CheckIfExistsInDB проверяет наличие записи в базе данных по shortURL
// true - если есть, false - если нет
func CheckIfExistsInDB(longURL string) (string, bool) {
	var existingURL string
	err := DB.QueryRow("SELECT short_url FROM urls WHERE original_url = $1", longURL).Scan(&existingURL)
	if err != nil {
		return "", false
	}
	return existingURL, true
}

// ReadFromDB reads a URL from the database
func ReadFromDB(shortURL string) (originalURL string, ok bool) {
	err := DB.QueryRow(selectSQL, shortURL).Scan(&originalURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read URL")
		return "", false
	}
	return originalURL, true
}

// BatchSave saves a batch of URLs to the database
func BatchSave(BatchURLs []BatchValues) {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Msg("Failed to commit transaction")
		}
	}()

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to prepare statement")
		return
	}
	defer stmt.Close()

	for _, v := range BatchURLs {
		_, err = stmt.Exec(v.UUID, v.ShortURL, v.OriginalURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to insert batch values")
			return
		}
	}
}
