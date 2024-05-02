package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"net/http"
	"strconv"
)

func (m *Repository) GetMenus(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := strconv.Atoi(chi.URLParam(r, "restaurant_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid restaurant id"))
		return
	}

	var menus []models.Menu
	err = m.App.DB.Where("restaurant_id = ?", restaurantID).Find(&menus).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  menus,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) GetMenu(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := strconv.Atoi(chi.URLParam(r, "restaurant_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid restaurant id"))
		return
	}

	menuID, err := strconv.Atoi(chi.URLParam(r, "menu_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu id"))
		return
	}

	var menu models.Menu
	err = m.App.DB.Where("restaurant_id = ? AND id = ?", restaurantID, menuID).First(&menu).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	var menuItems []models.MenuItem
	err = m.App.DB.Where("menu_id = ?", menu.ID).Find(&menuItems).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  menuItems,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) CreateMenu(w http.ResponseWriter, r *http.Request) {
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

	if !m.isAdmin(userID) {
		var restaurant models.Restaurant
		if err := m.App.DB.First(&restaurant, "id = ? AND owner_id = ?", restaurantID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	var newMenu models.Menu
	err = json.NewDecoder(r.Body).Decode(&newMenu)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding menu data"), http.StatusBadRequest)
		return
	}

	newMenu.RestaurantID = uint(restaurantID)

	if err := m.App.DB.Create(&newMenu).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  newMenu,
	}
	_ = m.writeJSON(w, http.StatusCreated, payload)
}

func (m *Repository) UpdateMenu(w http.ResponseWriter, r *http.Request) {
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

	menuID, err := strconv.Atoi(chi.URLParam(r, "menu_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu ID"))
		return
	}

	if !m.isAdmin(userID) {
		var existingRestaurant models.Restaurant
		if err := m.App.DB.First(&existingRestaurant, "id = ? AND owner_id = ?", restaurantID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	var existingMenu models.Menu
	if err := m.App.DB.First(&existingMenu, "id = ?", menuID).Error; err != nil {
		_ = m.errorJSON(w, errors.New("menu not found"), http.StatusNotFound)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&existingMenu)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding menu data"), http.StatusBadRequest)
		return
	}

	if err := m.App.DB.Save(&existingMenu).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  existingMenu,
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	userID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	menuID, err := strconv.Atoi(chi.URLParam(r, "menu_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu ID"), http.StatusBadRequest)
		return
	}

	var menu models.Menu
	if m.isAdmin(userID) {
		if err := m.App.DB.First(&menu, menuID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("menu not found"), http.StatusNotFound)
			return
		}
	} else {
		if err := m.App.DB.First(&menu, "id = ? AND restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = ?)", menuID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("menu not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	if err := m.App.DB.Delete(&menu).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "menu deleted successfully",
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}
