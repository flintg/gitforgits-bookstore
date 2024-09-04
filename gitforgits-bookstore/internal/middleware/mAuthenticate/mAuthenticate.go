package mAuthenticate

import (
	"log"
	"net/http"
)

/*
The AuthenticationMiddleware happens before the primary handler.
Comes from page 50 of Web Programming with Go; Building and Scaling Interactive Web Applications with Go's Robust Ecosystem by Ian Taylor, 2023 GitforGits
There is a little missing, here.
*/
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*if !userIsAuthenticated(r) {
			http.Redirect(w, r, "/login", 302)
			return
		}*/
		log.Println("AuthenticationMiddleware handler func served.")
		next.ServeHTTP(w, r)
	})
}
