package store

import (
	"database/sql"
	"errors"
	"github.com/rs/zerolog/log"
)

// DB a global variable to hold the database connection
var DB *sql.DB

// BatchValues a struct to hold the values for batch insert
type BatchValues struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UUID        string `json:"uuid"`
	DeletedFlag bool   `json:"-"`
}

// Define a struct to represent the row structure
type URLRow struct {
	OriginalURL string
	ShortURL    string
	UUID        string
	DeletedFlag bool
}

// SQL statement to create the table
const createTableSQL = `
		CREATE TABLE IF NOT EXISTS urls (
			uuid TEXT,
			short_url TEXT NOT NULL,
			original_url TEXT NOT NULL,
			deleted_flag BOOL NOT NULL DEFAULT FALSE
		);`

const softDeleteSQL = `
		UPDATE urls
		SET deleted_flag = true
		WHERE short_url = $1 AND uuid = $2;
	`

// SQL statement to insert into the table
const insertSQL = `INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3);`

// SQL statement to delete from the table
const deleteSQL = `DELETE FROM urls WHERE short_url = $1;`

// SQL statement to select from the table
const selectSQL = `SELECT uuid, short_url, original_url, deleted_flag FROM urls WHERE short_url = $1;`

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
func ReadFromDB(shortURL string) (URLRow, bool) {
	var urlRow URLRow
	err := DB.QueryRow(selectSQL, shortURL).Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL, &urlRow.DeletedFlag)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read URL")
		return URLRow{}, false
	}
	return urlRow, true
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

func GetAllURLsForUser(uuid string) ([]URLRow, error) {
	var urls []URLRow
	rows, err := DB.Query("SELECT uuid, short_url, original_url FROM urls WHERE uuid = $1", uuid)
	if err != nil || rows.Err() != nil { // А в чём разница?
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url URLRow
		err = rows.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func SoftDeleteRecord(shortURL, uuid string) error {
	record, ok := ReadFromDB(shortURL)
	if !ok {
		return errors.New("record not found")
	}
	if record.UUID != uuid {
		return errors.New("you do not have permission to delete this record")
	}
	_, err := DB.Exec(softDeleteSQL, shortURL, uuid)
	if err != nil {
		log.Err(err).Msgf("Error deleting record, %v", err)
		return err
	}
	log.Info().Msgf("Record with short URL %s deleted successfully\n", shortURL)
	return nil
}

func HardDeleteRecord() {
	_, err := DB.Exec("DELETE FROM urls WHERE deleted_flag = true")
	if err != nil {
		log.Error().Err(err).Msg("Failed to hard delete records")
	}
	log.Info().Msg("Records with deleted_flag=true deleted successfully")
}
