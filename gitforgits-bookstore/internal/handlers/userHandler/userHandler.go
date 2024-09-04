package userHandler

import (
	"fmt"
	_ "log"
	"net/http"

	"github.com/gorilla/mux"
)

var UserPathPrefix string = "/user"

/*
Handles user registration
*/
func UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	http.Error(w, fmt.Sprintf("Oops, UserRegisterHandler isn't implemented, yet. username [%v], email [%v], password [%v].", username, email, password), http.StatusNotImplemented)
}

/*
Handles user logins
*/
func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	http.Error(w, fmt.Sprintf("Oops, .userLoginHandler isn't implemented, yet. username [%v], password [%v]", username, password), http.StatusNotImplemented)
}

/*
Handles user profile actions
*/
func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops, .userProfileHandler isn't implemented, yet.", http.StatusNotImplemented)
}

/*
Handles 404 errors for user actions
*/
func UserNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("It's not you, it's us. We can't find what you're looking for."))
}

/*
Registers handlers and their subroutes. This function encapsulates the implementation of
the routes this package expects to handle.

Borrowed from StackOverflow answer https://stackoverflow.com/a/44391691, retrieved 2024-08-13
*/
func RegisterHandlers(r *mux.Router) {
	sr := r.PathPrefix(UserPathPrefix).Subrouter()
	sr.HandleFunc("/register", UserRegisterHandler)
	sr.HandleFunc("/login", UserLoginHandler)
	sr.HandleFunc("/profile", UserProfileHandler)
	sr.NotFoundHandler = http.HandlerFunc(UserNotFound)
}
