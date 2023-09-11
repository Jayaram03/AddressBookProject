package router

import (
	"go-postgres/middleware"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Router() http.Handler {

	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("frontend"))

	// Handle requests to the "/static/" path by serving static files
	router.PathPrefix("/frontend/").Handler(http.StripPrefix("/frontend/", fs))

	router.Use(corsMiddleware)
	// http.Handle("/", http.FileServer(http.Dir(".")))
	router.HandleFunc("/api/get/{id}", middleware.GetAddressbyId).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/getall", middleware.GetAlladdressBook).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/add", middleware.CreateAddress).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/update/{id}", middleware.UpdateAddress).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/delete/{id}", middleware.DeleteAddress).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/images/", middleware.ServeImage).Methods("GET", "OPTIONS")

	c := cors.Default()

	// Wrap the router with the CORS handler
	handler := c.Handler(router)

	return handler
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
