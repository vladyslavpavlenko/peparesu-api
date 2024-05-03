package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"net/http"
	"strconv"
)

func (m *Repository) GetMenuItem(w http.ResponseWriter, r *http.Request) {
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

	menuItemID, err := strconv.Atoi(chi.URLParam(r, "menu_item_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu item ID"))
		return
	}

	var menu models.Menu
	err = m.App.DB.Where("restaurant_id = ? AND id = ?", restaurantID, menuID).First(&menu).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	var menuItem models.MenuItem
	err = m.App.DB.Where("menu_id = ? AND id = ?", menu.ID, menuItemID).First(&menuItem).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  menuItem,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) LikeMenuItem(w http.ResponseWriter, r *http.Request) {
	menuID, err := strconv.Atoi(chi.URLParam(r, "menu_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu ID"), http.StatusBadRequest)
		return
	}

	menuItemID, err := strconv.Atoi(chi.URLParam(r, "menu_item_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu item ID"), http.StatusBadRequest)
		return
	}

	var menuItem models.MenuItem
	err = m.App.DB.Where("menu_id = ? AND id = ?", menuID, menuItemID).First(&menuItem).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusNotFound)
		return
	}

	action := chi.URLParam(r, "action")
	if action == "like" {
		menuItem.LikesCount++
	} else if action == "unlike" {
		if menuItem.LikesCount > 0 {
			menuItem.LikesCount--
		}
	} else {
		_ = m.errorJSON(w, errors.New("invalid action, expected 'like' or 'unlike'"), http.StatusBadRequest)
		return
	}

	err = m.App.DB.Save(&menuItem).Error
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  menuItem,
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	userID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := m.App.DB.First(&user, "id = ?", userID).Error; err != nil {
		_ = m.errorJSON(w, errors.New("user not found"), http.StatusUnauthorized)
		return
	}

	menuID, err := strconv.Atoi(chi.URLParam(r, "menu_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu ID"), http.StatusBadRequest)
		return
	}

	var menu models.Menu
	if err := m.App.DB.Joins("JOIN restaurants ON restaurants.id = menus.restaurant_id").
		First(&menu, "menus.id = ? AND (restaurants.owner_id = ? OR ? = 2)", menuID, userID, user.UserTypeID).Error; err != nil {
		if m.isAdmin(user.UserTypeID) {
			_ = m.errorJSON(w, errors.New("menu not found or access denied"), http.StatusNotFound)
		} else {
			_ = m.errorJSON(w, errors.New("menu not found"), http.StatusNotFound)
		}
		return
	}

	var newMenuItem models.MenuItem
	err = json.NewDecoder(r.Body).Decode(&newMenuItem)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding menu item data"), http.StatusBadRequest)
		return
	}

	newMenuItem.MenuID = uint(menuID)

	if err := m.App.DB.Create(&newMenuItem).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  newMenuItem,
	}
	_ = m.writeJSON(w, http.StatusCreated, payload)
}

func (m *Repository) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
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

	menuItemID, err := strconv.Atoi(chi.URLParam(r, "menu_item_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu item ID"))
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

	var existingMenuItem models.MenuItem
	if err := m.App.DB.First(&existingMenuItem, "id = ?", menuItemID).Error; err != nil {
		_ = m.errorJSON(w, errors.New("menu item not found"), http.StatusNotFound)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&existingMenuItem)
	if err != nil {
		_ = m.errorJSON(w, errors.New("error decoding menu data"), http.StatusBadRequest)
		return
	}

	if err := m.App.DB.Save(&existingMenuItem).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data:  existingMenuItem,
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}

func (m *Repository) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	userID, err := m.getUserFromToken(r)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	menuItemID, err := strconv.Atoi(chi.URLParam(r, "menu_item_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu item ID"), http.StatusBadRequest)
		return
	}

	var menuItem models.MenuItem
	if m.isAdmin(userID) {
		if err := m.App.DB.First(&menuItem, menuItemID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("menu item not found"), http.StatusNotFound)
			return
		}
	} else {
		if err := m.App.DB.First(&menuItem, "id = ? AND menu_id IN (SELECT id FROM menus WHERE restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = ?))", menuItemID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("menu item not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	if err := m.App.DB.Delete(&menuItem).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "menu item deleted successfully",
	}
	_ = m.writeJSON(w, http.StatusOK, payload)
}
