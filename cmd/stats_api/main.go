package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://torbnexus:secret@localhost:5433/torbnexus?sslmode=disable"
	}

	database, err := db.NewDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "https://tormentnexus.site")
		ctx := r.Context()

		companies, _ := database.CountCompanies(ctx)
		contacts, _ := database.CountContacts(ctx)
		interactions, _ := database.CountInteractions(ctx)
		stateCounts := make(map[string]int)
		states, _ := database.CountDealsByState(ctx)
		for _, st := range states {
			stateCounts[string(st.State)] = st.Count
		}

		data := map[string]interface{}{
			"companies": companies, "contacts": contacts,
			"interactions": interactions, "deals": stateCounts,
			"status": "operational",
		}
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/stats", http.StatusFound)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Stats API server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
