module golang-web-book/gitforgits-bookstore/internal/handlers/orderHandler

go 1.22.4

require (
	github.com/flintg/gitforgits-bookstore/mAuthenticate v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.1
)

replace github.com/flintg/gitforgits-bookstore/mAuthenticate => ../../middleware/mAuthenticate
