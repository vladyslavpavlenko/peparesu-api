package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vladyslavpavlenko/peparesu/config"
	"github.com/vladyslavpavlenko/peparesu/internal/handlers"
	"net/http"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Route("/api/v1", func(mux chi.Router) {
		// User
		// must but logged out
		mux.Group(func(mux chi.Router) {
			mux.Use(handlers.Repo.RequireNoAuth)

			mux.Post("/signup", handlers.Repo.SignUp)
			mux.Post("/login", handlers.Repo.Login)
		})
		// must but logged in
		mux.Group(func(mux chi.Router) {
			mux.Use(handlers.Repo.RequireAuth)

			mux.Post("/logout", handlers.Repo.Logout)

			mux.Post("/restaurants/create", handlers.Repo.CreateRestaurant)
			mux.Post("/restaurants/{restaurant_id}/menus/create", handlers.Repo.CreateMenu)
			mux.Post("/restaurants/{restaurant_id}/menus/{menu_id}/create", handlers.Repo.CreateMenuItem)

			mux.Put("/restaurants/{restaurant_id}/update", handlers.Repo.UpdateRestaurant)
			mux.Put("/restaurants/{restaurant_id}/menus/{menu_id}/update", handlers.Repo.UpdateMenu)
			mux.Put("/restaurants/{restaurant_id}/menus/{menu_id}/{menu_item_id}/update", handlers.Repo.UpdateMenuItem)

			mux.Delete("/restaurants/{restaurant_id}/delete", handlers.Repo.DeleteRestaurant)
			mux.Delete("/restaurants/{restaurant_id}/menus/{menu_id}/delete", handlers.Repo.DeleteMenu)
			mux.Delete("/restaurants/{restaurant_id}/menus/{menu_id}/{menu_item_id}/delete", handlers.Repo.DeleteMenuItem)
		})

		// Restaurant
		mux.Get("/restaurants", handlers.Repo.GetRestaurants)
		mux.Get("/restaurants/{restaurant_id}", handlers.Repo.GetRestaurant)

		// Menu
		mux.Get("/restaurants/{restaurant_id}/menus", handlers.Repo.GetMenus)
		mux.Get("/restaurants/{restaurant_id}/menus/{menu_id}", handlers.Repo.GetMenu)
		mux.Get("/restaurants/{restaurant_id}/menus/{menu_id}/{menu_item_id}", handlers.Repo.GetMenuItem)
		mux.Put("/restaurants/{restaurant_id}/menus/{menu_id}/{menu_item_id}/{action}", handlers.Repo.LikeMenuItem)
	})

	return mux
}
