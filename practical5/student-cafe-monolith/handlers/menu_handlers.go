package handlers

import (
	"encoding/json"
	"net/http"
	"student-cafe-monolith/database"
	"student-cafe-monolith/models"
)

func GetMenu(w http.ResponseWriter, r *http.Request) {
	var items []models.MenuItem
	if err := database.DB.Find(&items).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var item models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.DB.Create(&item).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}
