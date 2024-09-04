package orderHandler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/flintg/gitforgits-bookstore/mAuthenticate"
)

var OrderPathPrefix string = "/order"

func RegisterHandlers(r *mux.Router) {
	sr := r.PathPrefix(OrderPathPrefix).Subrouter()
	sr.HandleFunc("/{id:[0-9]+}", GetOrderDetail)
	sr.HandleFunc("/cart", GetOrdersCart)
	sr.HandleFunc("/checkout", GetOrdersCheckout)
	sr.HandleFunc("/history", GetOrdersHistory)
	sr.NotFoundHandler = http.HandlerFunc(OrderNotFound)
	sr.Use(mAuthenticate.AuthenticationMiddleware)
}

/*
Gets the detail of a single order
*/
func GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	//Handler logic to fetch and return book details
	vars := mux.Vars(r)
	orderID := vars["id"]
	http.Error(w, fmt.Sprintf("Oops, GetOrdersDetail isn't implemented, yet. id: [%v]", orderID), http.StatusNotImplemented)
}

/*
Handles cart actions
*/
func GetOrdersCart(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops, GetOrdersCart isn't implemented, yet.", http.StatusNotImplemented)
}

/*
Handles checkout actions
*/
func GetOrdersCheckout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops, GetOrdersCheckout isn't implemented, yet.", http.StatusNotImplemented)
}

/*
Handles order history actions
*/
func GetOrdersHistory(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops, GetOrdersHistory isn't implemented, yet.", http.StatusNotImplemented)
}

/*
Handles 404 errors for order actions
*/
func OrderNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("That's not in our Rolodex."))
}
