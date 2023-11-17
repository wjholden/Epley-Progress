package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

const url = `https://docs.google.com/spreadsheets/d/12DfaW5F8XM-DCjTMRgYZksnEyhKDShLYMETkhmhCspE/export?format=csv`

func main() {
	// Get data from Google Sheets
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Parse CSV
	r := csv.NewReader(resp.Body)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Open a SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(`
CREATE TABLE Epley
	(id INTEGER PRIMARY KEY AUTOINCREMENT,
		Lift TEXT NOT NULL,
		Weight REAL NOT NULL,
		Reps INTEGER NOT NULL,
		Estimate INTEGER AS (ROUND((0.033 * Reps * Weight) + Weight))
	)
`)
	defer db.Close()

	// Insert our data into the database
	for _, record := range records[1:] {
		//date := record[0]
		lift := record[1]
		weight := record[2]
		reps := record[3]
		query := fmt.Sprintf("INSERT INTO Epley (Lift, Weight, Reps) VALUES('%s', %s, %s)", lift, weight, reps)
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, lift := range []string{"Deadlift", "Squat", "Bench", "Press"} {
		getStats(lift, db)
		fmt.Println()
	}
}

func getStats(lift string, db *sql.DB) {
	fmt.Println(lift)
	query := fmt.Sprintf("SELECT Weight, Reps, Estimate FROM Epley WHERE Lift = '%s' ORDER BY id", lift)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal()
	}
	defer rows.Close()

	for rows.Next() {
		var reps, estimate int
		var weight float64
		err = rows.Scan(&weight, &reps, &estimate)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%3d kg for %2d => %3d\n", int(weight), reps, estimate)
	}
}
