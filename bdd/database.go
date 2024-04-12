package bdd

import (
	"database/sql"
	"log"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./groupie-tracker.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery1 := `
	CREATE TABLE IF NOT EXISTS USER (
		id INTEGER PRIMARY KEY,
		pseudo TEXT NOT NULL,
		email TEXT NOT NULL,
		password TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTableQuery1)
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery2 := `
	CREATE TABLE IF NOT EXISTS ROOMS (
		id INTEGER PRIMARY KEY,
		created_by INTEGER NOT NULL,
		max_player INTEGER NOT NULL,
		name TEXT NOT NULL,
		id_game INTEGER,
		FOREIGN KEY (created_by) REFERENCES USER(id),
		FOREIGN KEY (id_game) REFERENCES GAMES(id)
	);
	`
	_, err = db.Exec(createTableQuery2)
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery3 := `
	CREATE TABLE IF NOT EXISTS ROOM_USERS (
		id_room INTEGER,
		id_user INTEGER,
		score INTEGER,
		FOREIGN KEY (id_room) REFERENCES ROOMS(id),
		FOREIGN KEY (id_user) REFERENCES USER(id),
		PRIMARY KEY (id_room, id_user)
	);
	`
	_, err = db.Exec(createTableQuery3)
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery4 := `
	CREATE TABLE IF NOT EXISTS GAMES (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTableQuery4)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
