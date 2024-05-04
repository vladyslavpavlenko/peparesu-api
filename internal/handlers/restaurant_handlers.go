package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"net/http"
	"strconv"
)

func (m *Repository) GetRestaurants(w http.ResponseWriter, r *http.Request) {
	urlQuery := r.URL.Query()
	ownerID := urlQuery.Get("owner_id")

	var restaurants []models.Restaurant
	query := m.App.DB

	if ownerID != "" {
		id, err := strconv.Atoi(ownerID)
		if err != nil {
			_ = m.errorJSON(w, errors.New("invalid owner ID"), http.StatusNotFound)
			return
		}

		query = query.Where("owner_id = ?", id)
	}

	err := query.Preload("Owner").Find(&restaurants).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  restaurants,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := strconv.Atoi(chi.URLParam(r, "restaurant_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid restaurant ID"))
		return
	}

	var restaurant models.Restaurant
	err = m.App.DB.Preload("Owner").Where("id = ?", restaurantID).First(&restaurant).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  restaurant,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	ownerID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var newRestaurant models.Restaurant
	err = json.NewDecoder(r.Body).Decode(&newRestaurant)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding restaurant data"), http.StatusBadRequest)
		return
	}

	var existingRestaurant models.Restaurant
	result := m.App.DB.Where("LOWER(title) = LOWER(?) AND "+
		"LOWER(type) = LOWER(?) AND "+
		"LOWER(description) = LOWER(?) AND "+
		"LOWER(address) = LOWER(?) AND "+
		"phone = ?", newRestaurant.Title, newRestaurant.Type,
		newRestaurant.Description, newRestaurant.Address, newRestaurant.Phone).First(&existingRestaurant)
	if result.Error == nil {
		_ = m.errorJSON(w, errors.New("duplicate restaurant entry"), http.StatusConflict)
		return
	}

	newRestaurant.OwnerID = ownerID

	if err := m.App.DB.Create(&newRestaurant).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  newRestaurant,
	}
	_ = m.writeJSON(w, http.StatusCreated, payload)
}

func (m *Repository) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	userID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	restaurantID, err := strconv.Atoi(chi.URLParam(r, "restaurant_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid restaurant ID"))
		return
	}

	var updateData models.Restaurant
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding restaurant data"), http.StatusBadRequest)
		return
	}

	var existingRestaurant models.Restaurant
	if m.isAdmin(userID) {
		if err := m.App.DB.First(&existingRestaurant, "id = ?", restaurantID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found"), http.StatusNotFound)
			return
		}
	} else {
		if err := m.App.DB.First(&existingRestaurant, "id = ? AND owner_id = ?", restaurantID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	existingRestaurant.Title = updateData.Title
	existingRestaurant.Type = updateData.Type
	existingRestaurant.Description = updateData.Description
	existingRestaurant.Address = updateData.Address
	existingRestaurant.Phone = updateData.Phone

	if err := m.App.DB.Save(&existingRestaurant).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  existingRestaurant,
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) DeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	userID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	restaurantID, err := strconv.Atoi(chi.URLParam(r, "restaurant_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid restaurant ID"), http.StatusBadRequest)
		return
	}

	var restaurant models.Restaurant
	if m.isAdmin(userID) {
		if err := m.App.DB.First(&restaurant, restaurantID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found"), http.StatusNotFound)
			return
		}
	} else {
		if err := m.App.DB.First(&restaurant, "id = ? AND owner_id = ?", restaurantID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	if err := m.App.DB.Delete(&restaurant).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "restaurant deleted successfully",
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}
