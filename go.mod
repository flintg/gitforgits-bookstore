module github.com/flintg/gitforgits-bookstore/gitforgits-bookstore

go 1.22.4

require (
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
)

require github.com/joho/godotenv v1.5.1

require (
	github.com/flintg/gitforgits-bookstore/bookHandler v0.0.0-00010101000000-000000000000
	github.com/flintg/gitforgits-bookstore/userHandler v0.0.0-00010101000000-000000000000
	golang.org/x/time v0.5.0
)

require (
	github.com/flintg/gitforgits-bookstore/genreHandler v0.0.0-00010101000000-000000000000
	github.com/flintg/gitforgits-bookstore/orderHandler v0.0.0-00010101000000-000000000000
)

require github.com/flintg/gitforgits-bookstore/mAuthenticate v0.0.0-00010101000000-000000000000 // indirect

//replace github.com/flintg/gitforgits-bookstore/configHelper => ./gitforgits-bookstore/utils/configHelper
replace github.com/flintg/gitforgits-bookstore/userHandler => ./gitforgits-bookstore/internal/handlers/userHandler

replace github.com/flintg/gitforgits-bookstore/bookHandler => ./gitforgits-bookstore/internal/handlers/bookHandler

replace github.com/flintg/gitforgits-bookstore/orderHandler => ./gitforgits-bookstore/internal/handlers/orderHandler

replace github.com/flintg/gitforgits-bookstore/genreHandler => ./gitforgits-bookstore/internal/handlers/genreHandler

replace github.com/flintg/gitforgits-bookstore/mAuthenticate => ./gitforgits-bookstore/internal/middleware/mAuthenticate
