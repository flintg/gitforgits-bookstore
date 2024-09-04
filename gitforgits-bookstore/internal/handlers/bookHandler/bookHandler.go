package bookHandler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/flintg/gitforgits-bookstore/genreHandler"

	"github.com/gorilla/mux"
)

var BookPathPrefix string = "/book"
var TemplateDir string = "web/templates/"
var BookDetailTemplate string = "bookDetails.gohtml"
var templateCache *template.Template
var BookDB *sql.DB

type Book struct {
	ID          int
	Title       string
	Author      string
	Genre       int
	Description string
	ISBN        string
	Pages       int
	ImageURL    string
	Price       string
	//UserReview  string // this should be another struct or an array, probably
}

type BookHandler struct {
	Templates *template.Template //= template.New("").Delims("{{", "}}")
}

/*
Registers handlers and their subroutes. This function encapsulates the implementation of
the routes this package expects to handle.

Borrowed from StackOverflow answer https://stackoverflow.com/a/44391691, retrieved 2024-08-13
*/
func RegisterHandlers(r *mux.Router) {
	sr := r.PathPrefix(BookPathPrefix).Subrouter()
	sr.HandleFunc("/", GetBooks).Methods("GET")
	sr.HandleFunc("/add", AddBook)
	sr.HandleFunc("/{id:[0-9]+}", GetBookDetail).Methods("GET")
	sr.HandleFunc("/{id:[0-9]+}/update", UpdateBookDetail).Methods("PUT")
	sr.HandleFunc("/{id:[0-9]+}/delete", DeleteBook).Methods("DELETE")
	sr.HandleFunc("/{id:[0-9]+}/reviews", GetBookReviews).Methods("GET")
	sr.NotFoundHandler = http.HandlerFunc(GetBookNotFound)
}

/*
Gets a list of books
*/
func GetBooks(w http.ResponseWriter, r *http.Request) {
	//Handler logic to fetch and return book details
	var (
		fetchedBook  Book
		fetchedBooks []Book
		curQuery     string
		curSelect    string
		curWhere     string
	)
	// Genre_ID is an integer, so expect an integer. Failure means it's not an integer (insert: Mr. Burns emoji here)
	genre_id, err := strconv.Atoi(r.URL.Query().Get("genre"))
	if err == nil {
		//ToDo: Figure out a good way to handle "AND" and "OR" with multiple filter selections. But for now, we only expect one kind of filter, a genre.
		curWhere = strings.Join([]string{curWhere, "\"Genre_ID\"=", strconv.Itoa(genre_id)}, " ")
	}
	curSelect = "SELECT \"ID\",\"Title\",\"Author\",\"Genre_ID\",\"Description\",\"ISBN\",\"Price\" FROM \"Books\""
	if curWhere != "" {
		curWhere = strings.Join([]string{"WHERE", curWhere}, " ")
	}
	curQuery = strings.Join([]string{curSelect, curWhere, ";"}, " ")
	log.Printf("bookHandler.GetBooks; curQuery=[%v]", curQuery)
	rows, err := BookDB.Query(curQuery)
	if err != nil {
		//log.Printf()
		panic(fmt.Sprintf("bookHandler.GetBooks; BookDB.Query [%v] failed. Error: %v", curQuery, err))
		//http.Error(w, "", http.StatusInternalServerError)
	}
	defer rows.Close()
	if rows.Next() {
		keepLooping := true
		for keepLooping {
			err := rows.Err()
			if err != nil {
				log.Printf("bookHandler.GetBooks; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
				keepLooping = false
				http.Error(w, "No books found.", http.StatusNotFound)
				return
			}
			if err := rows.Scan(
				&fetchedBook.ID,
				&fetchedBook.Title,
				&fetchedBook.Author,
				&fetchedBook.Genre,
				&fetchedBook.Description,
				&fetchedBook.ISBN,
				&fetchedBook.Price,
			); err != nil {
				log.Printf("bookHandler.GetBooks; rows.Scan() on query [%v] failed. Error: %v", curQuery, err)
				http.Error(w, "", http.StatusInternalServerError)
				keepLooping = false
			} else {
				fetchedBooks = append(fetchedBooks, fetchedBook)
				keepLooping = rows.Next()
			}
		}
		if templateCache == nil {
			log.Print("bookHandler templateCache is nil.")
			panic("bookHandler.template is nil!")
		}
		err = templateCache.ExecuteTemplate(w, "bookList", fetchedBooks)
		if err != nil {
			log.Printf("bookHandler.GetBooks(w,r) error: %v", err)
		}
	} else {
		rows.Close()
		err := rows.Err()
		if err != nil {
			log.Printf("bookHandler.GetBooks; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
		}
		http.Error(w, "No books found.", http.StatusNotFound)
		return
	}
	//http.Error(w, fmt.Sprintf("Oops, .getBooks isn't implemented, yet. Query param genre=[%v]", genre), http.StatusNotImplemented)
}

/*
Adds new book to the catalogue
*/
func AddBook(w http.ResponseWriter, r *http.Request) {
	var (
		newBook  Book
		err      error
		curQuery string
	)
	rMethod := r.Method
	if rMethod == "" {
		rMethod = "GET"
	}
	switch rMethod {
	case "GET":
		var (
			curGenre  genreHandler.Genre
			allGenres []genreHandler.Genre
		)
		curQuery = "SELECT \"ID\",\"Name\" FROM \"Genres\""
		curQuery = fmt.Sprintf("%v;", curQuery)
		rows, queryErr := BookDB.Query(curQuery)
		if queryErr != nil {
			//log.Printf()
			panic(fmt.Sprintf("bookHandler.AddBook; BookDB.Query [%v] failed. Error: %v", curQuery, queryErr))
			//http.Error(w, "", http.StatusInternalServerError)
		}
		defer rows.Close()
		if rows.Next() {
			for {
				err = rows.Err()
				if err != nil {
					log.Printf("bookHandler.AddBook; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
					http.Error(w, "No books found.", http.StatusNotFound)
					break
				}
				if err = rows.Scan(
					&curGenre.ID,
					&curGenre.Name,
				); err != nil {
					log.Printf("bookHandler.AddBook; rows.Scan() on query [%v] failed. Error: %v", curQuery, err)
					http.Error(w, "", http.StatusInternalServerError)
					break
				} else {
					allGenres = append(allGenres, curGenre)
					if !rows.Next() {
						break
					}
				}
			}
			if err != nil {
				return
			}
		} else {
			rows.Close()
			err = rows.Err()
			if err != nil {
				log.Printf("bookHandler.GetBooks; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
			}
			http.Error(w, "No books found.", http.StatusNotFound)
			return
		}
		if templateCache == nil {
			log.Print("bookHandler templateCache is nil.")
			panic("bookHandler.template is nil!")
		}
		err = templateCache.ExecuteTemplate(w, "bookAdd", allGenres)
		if err != nil {
			log.Printf("bookHandler.AddBook(w,r) error: %v", err)
		}
	case "POST":
		rContentType := r.Header.Get("Content-Type")
		if rContentType == "application/json" {
			curBody, err := io.ReadAll(r.Body) // reads the request body stream into a variable that can be reused. Apparently the Body can only be read one time.
			if err != nil {
				log.Printf("bookHandler.AddBook; Error reading body. Error: %v", err)
				http.Error(w, "Could not read request.", http.StatusInternalServerError)
			}
			err = json.Unmarshal(curBody, &newBook) //json.NewDecoder(curBody) //converted to json.Unmarshal to leave work directly with the byte array from io.RadAll
			//err := decoder.Decode(&newBook)
			if err != nil {
				http.Error(w, "Invalid book data", http.StatusBadRequest)
				log.Printf("addBook: Bad request. Received \n%s", string(curBody)) // This probably isn't good in production
				return
			}
		} else if rContentType == "application/x-www-form-urlencoded" {
			err = r.ParseForm()
			if err != nil {
				http.Error(w, "Invalid form data.", http.StatusBadRequest)
				log.Printf("addBook: Error parsing form data. Error: %v", err)
				return
			}
			newBook.Title = r.FormValue("title")
			newBook.Author = r.FormValue("author")
			newBook.ISBN = r.FormValue("isbn")
			newBook.Description = r.FormValue("description")
			newBook.Price = r.FormValue("price")
			var (
				genre  int
				sGenre string
			)
			sGenre = r.FormValue("genre")
			log.Printf("New book to add: %v, %v, %v, %v, %v", newBook.Title, newBook.Author, newBook.ISBN, newBook.Description, sGenre)
			if genre, err = strconv.Atoi(sGenre); err == nil {
				newBook.Genre = genre
			} else {
				http.Error(w, "Genre must be an integer.", http.StatusBadRequest)
				log.Printf("addBook: Bad request. Genre must be an integer, received %s. Error: %v", sGenre, err)
				return
			}
		} else {
			http.Error(w, fmt.Sprintf("Unexpected Content-Type %s", rContentType), http.StatusBadRequest)
			log.Printf("addBook: Bad request. Unexpected Content-Type, received %s", rContentType)
			return
		}
		result, err := BookDB.Exec(
			"INSERT INTO \"Books\"(\"Title\",\"Author\",\"ISBN\",\"Description\",\"Genre_ID\",\"Price\") VALUES($1,$2,$3,$4,$5,$6)",
			newBook.Title, newBook.Author, newBook.ISBN, newBook.Description, newBook.Genre, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("bookHandler.AddBook error: %v", err), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(result)
		}
		/*BookstoreDB.Create($newBook)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&newBook)*/
	default:
		http.Error(w, fmt.Sprintf("Unsupported method %v", rMethod), http.StatusBadRequest)
		return
	}

}

/*
Gets the detail of a single book
*/
func GetBookDetail(w http.ResponseWriter, r *http.Request) {
	//Handler logic to fetch and return book details
	var (
		vars        = mux.Vars(r)
		bookID      = vars["id"]
		fetchedBook Book
		curQuery    string
	)
	curQuery = fmt.Sprintf("SELECT \"Title\",\"Author\",\"Genre_ID\",\"Description\",\"ISBN\",0 FROM \"Books\" WHERE \"ID\"=%v", bookID)
	rows, err := BookDB.Query(curQuery)
	if err != nil {
		//log.Printf()
		panic(fmt.Sprintf("bookHandler.GetBookDetail; BookDB.Query [%v] failed. Error: %v", curQuery, err))
		//http.Error(w, "", http.StatusInternalServerError)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&fetchedBook.Title, &fetchedBook.Author, &fetchedBook.Genre, &fetchedBook.Description, &fetchedBook.ISBN, &fetchedBook.Pages); err != nil {
			log.Printf("bookHandler.GetBookDetail; rows.Scan() on query [%v] failed. Error: %v", curQuery, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		if templateCache == nil {
			log.Print("bookHandler templateCache is nil.")
			panic("bookHandler.template is nil!")
		}
		err = templateCache.ExecuteTemplate(w, "bookDetails", fetchedBook)
		if err != nil {
			log.Printf("bookHandler.GetDetail(w,r) error: %v", err)
		}
	} else {
		rows.Close()
		err := rows.Err()
		if err != nil {
			log.Printf("bookHandler.GetBookDetail; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
		}
		http.Error(w, "Book not found.", http.StatusNotFound)
		return
	}
}

/*
Updates the detail of a single book
*/
func UpdateBookDetail(w http.ResponseWriter, r *http.Request) {
	var (
		vars        = mux.Vars(r)
		bookID      = vars["id"]
		updatedBook Book
	)
	/*if err := BookstoreDB.First(&updatedBook, bookID).Error; err != nil {
		http.Error(w, fmt.Sprintf("Book not found. [%v]", bookID), http.StatusNotFound)
		return
	}*/
	decoder := json.NewDecoder(r.Body)
	log.Printf(".updateBookDetail: received [%v]", r.Body)
	err := decoder.Decode(&updatedBook)
	log.Printf("Decoded: Title[%v], Author[%v], ISBN[%v], Pages[%v]", updatedBook.Title, updatedBook.Author, updatedBook.ISBN, updatedBook.Pages)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid data received for book [%v]", bookID), http.StatusBadRequest)
		return
	}
	// update the book in the DB
	/*
		BookstoreDB.Save(&updatedBook)
	*/
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&updatedBook)
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		bookID = vars["id"]
		//deletedBook Book
		curQuery   string
		dbResponse string
	)
	curQuery = fmt.Sprintf("DELETE FROM \"Books\" WHERE \"ID\"=%v", bookID)
	rows, err := BookDB.Query(curQuery)
	if err != nil {
		log.Printf("bookHandler.DeleteBook; BookDB.Query [%v] failed. Error: %v", curQuery, err)
		http.Error(w, "Unable to process the request.", http.StatusBadRequest)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&dbResponse); err != nil {
			log.Printf("bookHandler.DeleteBook; rows.Scan() on query [%v] failed. Error: %v", curQuery, err)
		} else {
			log.Printf("bookHandler.DeleteBook; Response: %v", dbResponse)
		}
	} else {
		rows.Close()
		err := rows.Err()
		if err != nil {
			log.Printf("bookHandler.DeleteBook; rows.Next() on query [%v] failed. Error: %v", curQuery, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
	/*if err := BookstoreDB.First(&deletedBook, bookID).Error; err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	BookstoreDB.Delete(&deletedBook)
	w.WriteHeader(http.StatusNoContent)
	*/
	//http.Error(w, fmt.Sprintf("Oops, .deleteBook isn't implemented, yet. bookID[%v]", bookID), http.StatusNotImplemented)
}

/*
Gets the reviews of a single book
*/
func GetBookReviews(w http.ResponseWriter, r *http.Request) {
	//Handler logic to fetch and return book reviews
	vars := mux.Vars(r)
	bookID := vars["id"]
	http.Error(w, fmt.Sprintf("Oops, .getBookReviews isn't implemented, yet. id: [%v]", bookID), http.StatusNotImplemented)
}

/*
404 handler for Books
*/
func GetBookNotFound(w http.ResponseWriter, r *http.Request) {
	//Handler logic to return 404 response
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Book not found."))
}

func (bh *BookHandler) LoadTemplates() {
	myTemplates, err := bh.Templates.ParseGlob("./*.gohtml")
	log.Println(fmt.Printf("myTemplates: %v\n", myTemplates))
	if err != nil {
		log.Fatalf("Could not load template [%v]. Error: %v", BookDetailTemplate, err)
	} else {
		log.Print("Book template loaded")
	}
	bh.Templates = myTemplates
	log.Println(fmt.Printf("templates: %v\n", bh.Templates))
}

func SetTemplateCache(t *template.Template) {
	templateCache = t
}
