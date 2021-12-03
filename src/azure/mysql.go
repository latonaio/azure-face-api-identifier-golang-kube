package azure

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func GetGuestID(db *sql.DB, personID string) int {

	// MySQL内にデータがあるか確認
	rows := db.QueryRow("SELECT guest_id FROM guest where face_id_azure = ?", personID)

	var guestIDFromMYSQL int
	rows.Scan(&guestIDFromMYSQL)

	return guestIDFromMYSQL
}
