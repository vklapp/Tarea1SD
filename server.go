package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"


	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Entidades
type Driver struct {
	DriverNumber int    `json:"driver_number"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	NameAcronym  string `json:"name_acronym"`
	TeamName     string `json:"team_name"`
	CountryCode  string `json:"country_code"`
}

type Session struct {
	SessionKey        int    `json:"session_key"`
	SessionName       string `json:"session_name"`
	SessionType       string `json:"session_type"`
	Location          string `json:"location"`
	CountryName       string `json:"country_name"`
	Year              int    `json:"year"`
	CircuitShortName  string `json:"circuit_short_name"`
	DateStart         string `json:"date_start"`
}

type Lap struct {
	DriverNumber     int     `json:"driver_number"`
	SessionKey       int     `json:"session_key"`
	LapNumber        int     `json:"lap_number"`
	LapDuration      float64 `json:"lap_duration"`
	DurationSector1  float64 `json:"duration_sector_1"`
	DurationSector2  float64 `json:"duration_sector_2"`
	DurationSector3  float64 `json:"duration_sector_3"`
	StSpeed          float64 `json:"st_speed"`
	DateStart        string  `json:"date_start"`
}

type Position struct {
	DriverNumber int    `json:"driver_number"`
	SessionKey   int    `json:"session_key"`
	Position     int    `json:"position"`
	Date         string `json:"date"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "/home/ubuntu/proxydb_mount/proxy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	cargarDatosDesdeOpenF1()
	r := gin.Default()

	// Endpoints
	r.GET("/api/corredor", getDrivers)
	r.GET("/api/corredor/detalle/:id", getDriverDetail)
	r.GET("/api/carrera", getCarreras)
	r.GET("/api/carrera/detalle/:id", getCarreraDetail)
	r.GET("/api/temporada/resumen", getResumenTemporada)

	r.Run(":8080")
}

// ---------------- ENDPOINTS ----------------

func getDrivers(c *gin.Context) {
	rows, _ := db.Query("SELECT driver_number, first_name, last_name, name_acronym, team_name, country_code FROM drivers")
	defer rows.Close()
	var list []Driver
	for rows.Next() {
		var d Driver
		rows.Scan(&d.DriverNumber, &d.FirstName, &d.LastName, &d.NameAcronym, &d.TeamName, &d.CountryCode)
		list = append(list, d)
	}
	c.JSON(200, list)
}

func getDriverDetail(c *gin.Context) {
	id := c.Param("id")
	driverID := id
	rows, _ := db.Query(`
		SELECT s.session_key, s.circuit_short_name, s.session_name, p.position, MAX(l.st_speed), MIN(l.lap_duration)
		FROM positions p
		JOIN sessions s ON p.session_key = s.session_key
		JOIN laps l ON l.session_key = p.session_key AND l.driver_number = p.driver_number
		WHERE p.driver_number = ?
		GROUP BY s.session_key
	`, driverID)
	defer rows.Close()

	type RaceResult struct {
		SessionKey       int     `json:"session_key"`
		CircuitShortName string  `json:"circuit_short_name"`
		Race             string  `json:"race"`
		Position         int     `json:"position"`
		FastestLap       bool    `json:"fastest_lap"`
		MaxSpeed         float64 `json:"max_speed"`
		BestLapDuration  float64 `json:"best_lap_duration"`
	}
	var results []RaceResult
	var wins, top3 int
	var maxSpeed float64
	for rows.Next() {
		var r RaceResult
		rows.Scan(&r.SessionKey, &r.CircuitShortName, &r.Race, &r.Position, &r.MaxSpeed, &r.BestLapDuration)
		if r.Position == 1 {
			wins++
		}
		if r.Position <= 3 {
			top3++
		}
		if r.MaxSpeed > maxSpeed {
			maxSpeed = r.MaxSpeed
		}
		r.FastestLap = true // simplificado
		results = append(results, r)
	}

	c.JSON(200, gin.H{
		"driver_id": driverID,
		"performance_summary": gin.H{
			"wins":      wins,
			"top_3_finishes": top3,
			"max_speed": maxSpeed,
		},
		"race_results": results,
	})
}

func getCarreras(c *gin.Context) {
	rows, _ := db.Query(`
		SELECT session_key, country_name, date_start, year, circuit_short_name
		FROM sessions
		WHERE session_type = 'Race'
	`)
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id, year int
		var pais, fecha, circuito string
		rows.Scan(&id, &pais, &fecha, &year, &circuito)
		list = append(list, gin.H{
			"session_key":       id,
			"country_name":      pais,
			"date_start":        fecha,
			"year":              year,
			"circuit_short_name": circuito,
		})
	}
	c.JSON(200, list)
}

func getCarreraDetail(c *gin.Context) {
	id := c.Param("id")
	sessionID := id

	// Resultados de la carrera
	rows, _ := db.Query(`
		SELECT p.position, d.first_name || ' ' || d.last_name, d.team_name, d.country_code
		FROM positions p
		JOIN drivers d ON d.driver_number = p.driver_number
		WHERE p.session_key = ?
		ORDER BY p.position ASC
	`, sessionID)
	defer rows.Close()
	var podio []gin.H
	var ultimo gin.H
	for rows.Next() {
		var pos int
		var nombre, equipo, pais string
		rows.Scan(&pos, &nombre, &equipo, &pais)
		dato := gin.H{
			"position": pos,
			"driver":   nombre,
			"team":     equipo,
			"country":  pais,
		}
		if pos <= 3 {
			podio = append(podio, dato)
		}
		ultimo = dato
	}
	// Vuelta rápida
	lapRow := db.QueryRow(`
		SELECT d.first_name || ' ' || d.last_name, l.lap_duration, l.duration_sector_1, l.duration_sector_2, l.duration_sector_3
		FROM laps l
		JOIN drivers d ON d.driver_number = l.driver_number
		WHERE l.session_key = ?
		ORDER BY l.lap_duration ASC
		LIMIT 1
	`, sessionID)
	var piloto string
	var total, s1, s2, s3 float64
	lapRow.Scan(&piloto, &total, &s1, &s2, &s3)

	// Velocidad máxima
	speedRow := db.QueryRow(`
		SELECT d.first_name || ' ' || d.last_name, MAX(l.st_speed)
		FROM laps l
		JOIN drivers d ON d.driver_number = l.driver_number
		WHERE l.session_key = ?
	`, sessionID)
	var pilotoVel string
	var vmax float64
	speedRow.Scan(&pilotoVel, &vmax)

	c.JSON(200, gin.H{
		"race_id": sessionID,
		"results": append(podio, gin.H{"position": "Ultimo", "driver": ultimo["driver"], "team": ultimo["team"], "country": ultimo["country"]}),
		"fastest_lap": gin.H{
			"driver":    piloto,
			"total_time": total,
			"sector_1": s1,
			"sector_2": s2,
			"sector_3": s3,
		},
		"max_speed": gin.H{
			"driver":    pilotoVel,
			"speed_kmh": vmax,
		},
	})
}

func getResumenTemporada(c *gin.Context) {
	// Victorias
	victorias := make(map[string]int)
	rows, _ := db.Query("SELECT d.first_name || ' ' || d.last_name, COUNT(*) FROM positions p JOIN drivers d ON p.driver_number = d.driver_number WHERE p.position = 1 GROUP BY d.driver_number")
	for rows.Next() {
		var nombre string
		var count int
		rows.Scan(&nombre, &count)
		victorias[nombre] = count
	}

	// Vueltas rápidas
	vueltasRapidas := make(map[string]int)
	rows, _ = db.Query("SELECT d.first_name || ' ' || d.last_name, COUNT(*) FROM laps l JOIN drivers d ON d.driver_number = l.driver_number WHERE l.lap_duration = (SELECT MIN(l2.lap_duration) FROM laps l2 WHERE l2.session_key = l.session_key) GROUP BY d.driver_number")
	for rows.Next() {
		var nombre string
		var count int
		rows.Scan(&nombre, &count)
		vueltasRapidas[nombre] = count
	}

	// Pole positions (simplificado como posición = 1 en primera fecha)
	poles := make(map[string]int)
	rows, _ = db.Query("SELECT d.first_name || ' ' || d.last_name, COUNT(*) FROM positions p JOIN drivers d ON p.driver_number = d.driver_number WHERE p.position = 1 GROUP BY p.driver_number")
	for rows.Next() {
		var nombre string
		var count int
		rows.Scan(&nombre, &count)
		poles[nombre] = count
	}

	// Convertir y ordenar
	type Stat struct {
		Position int    `json:"position"`
		Driver   string `json:"driver"`
		Value    int    `json:"wins,omitempty" json:"fastest_laps,omitempty" json:"poles,omitempty"`
	}

	getTop := func(m map[string]int) []Stat {
		var stats []Stat
		for k, v := range m {
			stats = append(stats, Stat{Driver: k, Value: v})
		}
		sort.Slice(stats, func(i, j int) bool { return stats[i].Value > stats[j].Value })
		for i := range stats {
			stats[i].Position = i + 1
		}
		if len(stats) > 3 {
			return stats[:3]
		}
		return stats
	}

	c.JSON(200, gin.H{
		"season":             2024,
		"top_3_winners":      getTop(victorias),
		"top_3_fastest_laps": getTop(vueltasRapidas),
		"top_3_pole_positions": getTop(poles),
	})
}

// ---------- CARGA DE DATOS DESDE OPENF1 -----------

func cargarDatosDesdeOpenF1() {
	log.Println("Cargando datos desde OpenF1...")

	cargarPilotos()
	cargarSesiones()
	cargarPosiciones()
	cargarVueltas()

	log.Println("Datos cargados correctamente.")
}

func cargarPilotos() {
	type Driver struct {
		DriverNumber int    `json:"driver_number"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		NameAcronym  string `json:"name_acronym"`
		TeamName     string `json:"team_name"`
		CountryCode  string `json:"country_code"`
	}

	insert := `INSERT OR IGNORE INTO drivers (driver_number, first_name, last_name, name_acronym, team_name, country_code) VALUES (?, ?, ?, ?, ?, ?)`

	sessions := map[int][]int{
		9574: {1, 2, 3, 4, 10, 11, 14, 16, 18, 20, 22, 23, 24, 27, 31, 44, 55, 63, 77, 81},
		9636: {30, 50, 43},
	}

	for sessionKey, permitidos := range sessions {
		url := fmt.Sprintf("https://api.openf1.org/v1/drivers?session_key=%d", sessionKey)
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error al obtener pilotos:", err)
			continue
		}
		defer resp.Body.Close()

		var drivers []Driver
		if err := json.NewDecoder(resp.Body).Decode(&drivers); err != nil {
			log.Println("Error decodificando pilotos:", err)
			continue
		}

		for _, d := range drivers {
			if contains(permitidos, d.DriverNumber) {
				_, _ = db.Exec(insert, d.DriverNumber, d.FirstName, d.LastName, d.NameAcronym, d.TeamName, d.CountryCode)
			}
		}
	}
}

func cargarSesiones() {
	url := "https://api.openf1.org/v1/sessions?session_name=Race&year=2024"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error al obtener sesiones:", err)
		return
	}
	defer resp.Body.Close()

	type Session struct {
		SessionKey       int    `json:"session_key"`
		SessionName      string `json:"session_name"`
		SessionType      string `json:"session_type"`
		Location         string `json:"location"`
		CountryName      string `json:"country_name"`
		Year             int    `json:"year"`
		CircuitShortName string `json:"circuit_short_name"`
		DateStart        string `json:"date_start"`
	}

	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		log.Println("Error decodificando sesiones:", err)
		return
	}

	insert := `INSERT OR IGNORE INTO sessions (session_key, session_name, session_type, location, country_name, year, circuit_short_name, date_start) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	for _, s := range sessions {
		_, _ = db.Exec(insert, s.SessionKey, s.SessionName, s.SessionType, s.Location, s.CountryName, s.Year, s.CircuitShortName, s.DateStart)
	}
}

func cargarPosiciones() {
	rows, _ := db.Query("SELECT session_key FROM sessions")
	var sessionKeys []int
	for rows.Next() {
		var k int
		rows.Scan(&k)
		sessionKeys = append(sessionKeys, k)
	}

	type Position struct {
		DriverNumber int    `json:"driver_number"`
		SessionKey   int    `json:"session_key"`
		Position     int    `json:"position"`
		Date         string `json:"date"`
	}

	insert := `INSERT OR IGNORE INTO positions (driver_number, session_key, position, date) VALUES (?, ?, ?, ?)`

	for _, key := range sessionKeys {
		url := fmt.Sprintf("https://api.openf1.org/v1/position?session_key=%d", key)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var positions []Position
		if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
			continue
		}

		for _, p := range positions {
			_, _ = db.Exec(insert, p.DriverNumber, p.SessionKey, p.Position, p.Date)
		}
	}
}

func cargarVueltas() {
	rows, _ := db.Query("SELECT session_key FROM sessions")
	var sessionKeys []int
	for rows.Next() {
		var k int
		rows.Scan(&k)
		sessionKeys = append(sessionKeys, k)
	}

	type Lap struct {
		DriverNumber     int     `json:"driver_number"`
		SessionKey       int     `json:"session_key"`
		LapNumber        int     `json:"lap_number"`
		LapDuration      float64 `json:"lap_duration"`
		DurationSector1  float64 `json:"duration_sector_1"`
		DurationSector2  float64 `json:"duration_sector_2"`
		DurationSector3  float64 `json:"duration_sector_3"`
		StSpeed          float64 `json:"st_speed"`
		DateStart        string  `json:"date_start"`
	}

	insert := `INSERT OR IGNORE INTO laps (driver_number, session_key, lap_number, lap_duration, duration_sector_1, duration_sector_2, duration_sector_3, st_speed, date_start) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, key := range sessionKeys {
		url := fmt.Sprintf("https://api.openf1.org/v1/laps?session_key=%d", key)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var laps []Lap
		if err := json.NewDecoder(resp.Body).Decode(&laps); err != nil {
			continue
		}

		for _, l := range laps {
			_, _ = db.Exec(insert, l.DriverNumber, l.SessionKey, l.LapNumber, l.LapDuration, l.DurationSector1, l.DurationSector2, l.DurationSector3, l.StSpeed, l.DateStart)
		}
	}
}

// Utilidad para verificar si un número está en una lista
func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
