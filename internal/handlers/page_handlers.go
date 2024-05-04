package handlers

import (
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"github.com/vladyslavpavlenko/peparesu/internal/render"
	"net/http"
)

func (m *Repository) Restaurants(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "restaurants.page.gohtml", &models.TemplateData{})
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}

func (m *Repository) Restaurant(w http.ResponseWriter, r *http.Request) {
	err := render.Template(w, r, "restaurant.page.gohtml", &models.TemplateData{})
	if err != nil {
		_ = m.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}
