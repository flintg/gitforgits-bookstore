package genreHandler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var GenrePathPrefix string = "/genre"
var templateCache *template.Template
var GenreDB *sql.DB

type Genre struct {
	ID   int
	Name string
}

type GenreHandler struct {
	Templates *template.Template //= template.New("").Delims("{{", "}}")
}

/*
Registers handlers and their subroutes. This function encapsulates the implementation of
the routes this package expects to handle.

Borrowed from StackOverflow answer https://stackoverflow.com/a/44391691, retrieved 2024-08-13
*/
func RegisterHandlers(r *mux.Router) {
	sr := r.PathPrefix(GenrePathPrefix).Subrouter()
	sr.HandleFunc("/", GetGenres).Methods("GET")
	sr.HandleFunc("/add", AddGenre)
	sr.HandleFunc("/{id:[0-9]+}", GetGenreDetail).Methods("GET")
	sr.HandleFunc("/{id:[0-9]+}/update", UpdateGenreDetail).Methods("PUT")
	sr.HandleFunc("/{id:[0-9]+}/delete", DeleteGenre).Methods("POST", "DELETE")
	sr.NotFoundHandler = http.HandlerFunc(GetGenreNotFound)
}

func GetGenres(w http.ResponseWriter, r *http.Request) {
	//Handler logic to fetch and return list of books
	genre := r.URL.Query().Get("genre")
	//Handler logic to fetch and return book details
	var (
		fetchedGenre  Genre
		fetchedGenres []Genre
		curQuery      string
		err           error
	)
	curQuery = "SELECT \"ID\",\"Name\" FROM \"Genres\""
	if genre != "" {
		curQuery = fmt.Sprintf(" WHERE \"ID\"=%v", genre)
	}
	curQuery = fmt.Sprintf("%v;", curQuery)
	rows, queryErr := GenreDB.Query(curQuery)
	if queryErr != nil {
		//log.Printf()
		panic(fmt.Sprintf("genreHandler.GetGenres; GenreDB.Query [%v] failed. Error: %v", curQuery, queryErr))
		//http.Error(w, "", http.StatusInternalServerError)
	}
	defer rows.Close()
	if rows.Next() {
		for {
			err = rows.Err()
			if err != nil {
				log.Printf("bookHandler.GetBooks; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
				http.Error(w, "No books found.", http.StatusNotFound)
				break
			}
			if err = rows.Scan(
				&fetchedGenre.ID,
				&fetchedGenre.Name,
			); err != nil {
				log.Printf("genreHandler.GetGenres; rows.Scan() on query [%v] failed. Error: %v", curQuery, err)
				http.Error(w, "", http.StatusInternalServerError)
				break
			} else {
				fetchedGenres = append(fetchedGenres, fetchedGenre)
				if !rows.Next() {
					break
				}
			}
		}
		if err != nil {
			return
		}
		if templateCache == nil {
			log.Print("genreHandler templateCache is nil.")
			panic("genreHandler.template is nil!")
		}
		err = templateCache.ExecuteTemplate(w, "genreList", fetchedGenres)
		if err != nil {
			log.Printf("genreHandler.GetGenres(w,r) error: %v", err)
		}
	} else {
		rows.Close()
		err := rows.Err()
		if err != nil {
			log.Printf("genreHandler.GetGenres; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
		}
		http.Error(w, "No books found.", http.StatusNotFound)
		return
	}
}

/*
Adds new genre to the catalogue
*/
func AddGenre(w http.ResponseWriter, r *http.Request) {
	var (
		newGenre Genre
		err      error
	)
	rMethod := r.Method
	if rMethod == "" {
		rMethod = "GET"
	}
	switch rMethod {
	case "GET":
		if templateCache == nil {
			log.Print("genreHandler templateCache is nil.")
			panic("genreHandler.template is nil!")
		}
		err = templateCache.ExecuteTemplate(w, "genreAdd", nil)
		if err != nil {
			log.Printf("genreHandler.AddGenre(w,r) error: %v", err)
		}
		return
	case "POST":
		rContentType := r.Header.Get("Content-Type")
		switch rContentType {
		case "application/json":
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&newGenre)
			if err != nil {
				http.Error(w, "Invalid genre data", http.StatusBadRequest)
				log.Printf("addGenre: Bad request. Received [%v]", r.Body) // This probably isn't good in production
				return
			}
		case "application/x-www-form-urlencoded":
			err = r.ParseForm()
			if err != nil {
				http.Error(w, "Invalid form data.", http.StatusBadRequest)
				log.Printf("addGenre: Error parsing form data. Error: %v", err)
				return
			}
			newGenre.Name = r.FormValue("name")
		default:
			http.Error(w, fmt.Sprintf("Unexpected Content-Type %s", rContentType), http.StatusBadRequest)
			log.Printf("addGenre: Bad request. Unexpected Content-Type, received %s", rContentType)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("Unsupported method %v", rMethod), http.StatusBadRequest)
		return
	}
	result, err := GenreDB.Exec(
		"INSERT INTO \"Genres\"(\"Name\") VALUES($1)",
		newGenre.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("genreHandler.AddGenre error: %v", err), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	}
	/*BookstoreDB.Create($newBook)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&newBook)*/
}

func GetGenreDetail(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ".GetGenreDetail method not implemented.", http.StatusNotImplemented)
}

func UpdateGenreDetail(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ".UpdateGenreDetail method not implemented.", http.StatusNotImplemented)
}

func DeleteGenre(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ".DeleteGenre method not implemented.", http.StatusNotImplemented)
}

func GetGenreNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ".GetGenreNotFound method not implemented.", http.StatusNotImplemented)
}

func SetTemplateCache(t *template.Template) {
	templateCache = t
}
