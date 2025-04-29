package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "proxy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables(db)
	log.Println("Base de datos creada correctamente.")
}

func createTables(db *sql.DB) {
	queries := []string{
		// Tabla de pilotos
		`CREATE TABLE IF NOT EXISTS drivers (
			driver_number INTEGER PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			name_acronym TEXT,
			team_name TEXT,
			country_code TEXT
		);`,

		// Tabla de sesiones
		`CREATE TABLE IF NOT EXISTS sessions (
			session_key INTEGER PRIMARY KEY,
			session_name TEXT,
			session_type TEXT,
			location TEXT,
			country_name TEXT,
			year INTEGER,
			circuit_short_name TEXT,
			date_start TEXT
		);`,

		// Tabla de posiciones
		`CREATE TABLE IF NOT EXISTS positions (
			driver_number INTEGER,
			session_key INTEGER,
			position INTEGER,
			date TEXT,
			PRIMARY KEY(driver_number, session_key)
		);`,

		// Tabla de vueltas
		`CREATE TABLE IF NOT EXISTS laps (
			driver_number INTEGER,
			session_key INTEGER,
			lap_number INTEGER,
			lap_duration REAL,
			duration_sector_1 REAL,
			duration_sector_2 REAL,
			duration_sector_3 REAL,
			st_speed REAL,
			date_start TEXT,
			PRIMARY KEY(driver_number, session_key, lap_number)
		);`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Fatalf("Error creando tabla: %v", err)
		}
	}
}
