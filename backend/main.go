package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type ContainerStatus struct {
	ID          int       `json:"id"`
	IPAddress   string    `json:"ip_address"`
	PingStatus  bool      `json:"ping_status"`
	LastChecked time.Time `json:"last_checked"`
}

var db *sql.DB

func createTableIfNotExists() {
	query := `
	CREATE TABLE IF NOT EXISTS container_status (
		id SERIAL PRIMARY KEY,
		ip_address VARCHAR(255) NOT NULL,
		ping_status BOOLEAN NOT NULL,
		last_checked TIMESTAMPTZ NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table: ", err)
	}
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgres://vk:password@db:5432/containers?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	createTableIfNotExists()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/data", handleData)

	handler := c.Handler(mux)

	log.Println("Backend running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func handleData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, ip_address, ping_status, last_checked FROM container_status")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var statuses []ContainerStatus
		for rows.Next() {
			var status ContainerStatus
			if err := rows.Scan(&status.ID, &status.IPAddress, &status.PingStatus, &status.LastChecked); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			statuses = append(statuses, status)
		}
		json.NewEncoder(w).Encode(statuses)

	case http.MethodPost:
		var status ContainerStatus
		if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec(
			"INSERT INTO container_status (ip_address, ping_status, last_checked) VALUES ($1, $2, $3)",
			status.IPAddress, status.PingStatus, time.Now(),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
