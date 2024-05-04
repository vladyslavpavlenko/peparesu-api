package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

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
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		_ = m.errorJSON(w, errors.New("error parsing form"), http.StatusBadRequest)
		return
	}

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
	newMenuItem.Title = r.FormValue("title")
	newMenuItem.Description = r.FormValue("description")
	newMenuItem.MenuID = uint(menuID)
	newMenuItem.Picture = "http://localhost:8080/api/v1/storage/images/menuitem-default.jpeg"

	priceUAH, _ := strconv.ParseInt(r.FormValue("price_uah"), 10, 64)
	newMenuItem.PriceUAH = uint(priceUAH)

	if err := m.App.DB.Create(&newMenuItem).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("picture")
	if err != nil {
		_ = m.errorJSON(w, errors.New("error retrieving the file"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := filepath.Join("storage/images", fmt.Sprintf("menuitem-%d.jpeg", newMenuItem.ID))
	dst, err := os.Create(filePath)
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	newMenuItem.Picture = fmt.Sprintf("http://localhost:8080/api/v1/%s", filePath)

	if err := m.App.DB.Save(&newMenuItem).Error; err != nil {
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
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		_ = m.errorJSON(w, errors.New("error parsing form"), http.StatusBadRequest)
		return
	}

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

	menuItemID, err := strconv.Atoi(chi.URLParam(r, "menu_item_id"))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid menu item ID"), http.StatusBadRequest)
		return
	}

	if !m.isAdmin(userID) {
		var existingRestaurant models.Restaurant
		if err := m.App.DB.First(&existingRestaurant, "id = ? AND owner_id = ?", restaurantID, userID).Error; err != nil {
			_ = m.errorJSON(w, errors.New("restaurant not found or not owned by the user"), http.StatusNotFound)
			return
		}
	}

	var existingMenuItem models.MenuItem
	if err := m.App.DB.First(&existingMenuItem, "id = ?", menuItemID).Error; err != nil {
		_ = m.errorJSON(w, errors.New("menu item not found"), http.StatusNotFound)
		return
	}

	existingMenuItem.Title = r.FormValue("title")
	existingMenuItem.Description = r.FormValue("description")

	priceUAH, _ := strconv.ParseInt(r.FormValue("price_uah"), 10, 64)
	existingMenuItem.PriceUAH = uint(priceUAH)

	file, _, err := r.FormFile("picture")
	if err == nil {
		defer file.Close()

		filePath := filepath.Join("storage/images", fmt.Sprintf("menuitem-%d.jpeg", menuItemID))
		dst, err := os.Create(filePath)
		if err != nil {
			_ = m.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			_ = m.errorJSON(w, err, http.StatusInternalServerError)
			return
		}

		existingMenuItem.Picture = fmt.Sprintf("http://localhost:8080/api/v1/%s", filePath)
	} else if !errors.Is(err, http.ErrMissingFile) {
		_ = m.errorJSON(w, errors.New("error processing uploaded file"), http.StatusBadRequest)
		return
	}

	if err := m.App.DB.Save(&existingMenuItem).Error; err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	_ = m.writeJSON(w, http.StatusOK, jsonResponse{
		Error: false,
		Data:  existingMenuItem,
	})
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
