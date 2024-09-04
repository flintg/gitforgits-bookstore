package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flintg/gitforgits-bookstore/bookHandler"
	"github.com/flintg/gitforgits-bookstore/genreHandler"
	"github.com/flintg/gitforgits-bookstore/orderHandler"
	"github.com/flintg/gitforgits-bookstore/userHandler"

	//_ "github.com/flintg/gitforgits-bookstore/configHelper" // This isn't working. Review https://go.dev/doc/tutorial/create-module

	// the underscore preceding this import is critical; we want to import the pSQL package even though we never directly reference it
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"

	"os"

	"github.com/joho/godotenv"

	"golang.org/x/time/rate"
)

type App struct {
	Configs      Cfg
	Router       *mux.Router
	StaticRouter *mux.Router
	DB           *sql.DB
}

type Cfg struct {
	ServerAddress string
	DbAddress     string
	DbUser        string
	DbPassword    string
	DbHost        string
	DbName        string
}

type HomeTemplate struct {
	PageTitle string
	PageBody  string
}

var limiter = rate.NewLimiter(5, 1)
var TemplateCache = template.New("").Delims("{{", "}}")

func (cfg *Cfg) Load() {

	//Load .env file during local development
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file, assuming production environment with OS level environment variables. Error: %v", err)
	}
	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}
	cfg.DbAddress = os.Getenv("DB_ADDRESS")
	cfg.DbUser = os.Getenv("DB_USER")
	cfg.DbPassword = os.Getenv("DB_PASSWORD")
	cfg.DbHost = os.Getenv("DB_HOST")
	cfg.DbName = os.Getenv("DB_NAME")
}

func (a *App) Initialize() {
	//Database connection logic
	a.Configs.Load()
	connectionString := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", a.Configs.DbUser, a.Configs.DbName, a.Configs.DbPassword, a.Configs.DbHost)
	log.Printf("Initialize(), connectionString = [%v]", strings.ReplaceAll(connectionString, a.Configs.DbPassword, "***"))
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	} else {
		err = a.DB.Ping()
		if err != nil {
			log.Printf("App.Initialize(); connection not open? Error: %v", err)
		} else {
			bookHandler.BookDB = a.DB
			genreHandler.GenreDB = a.DB
			log.Print("We have a connection to the database.")
		}
	}
	a.DB.SetMaxOpenConns(100)
	a.DB.SetMaxIdleConns(50)
	a.DB.SetConnMaxLifetime(time.Minute * 5)
	//Initialize Router and Routes
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

/*
Initializes the routes. To add a new route in the application,
modify this function to reflect the new route.

Gorilla Mux handles routes
from top to bottom, stopping at the first matched route. Keep a catch-all
route at the end of the list of routes ("/").
*/
func (a *App) initializeRoutes() {
	//Book routing
	bookHandler.BookPathPrefix = "/books" //default is /book (singular)
	bookHandler.RegisterHandlers(a.Router)
	//User routing
	userHandler.RegisterHandlers(a.Router)
	//Order routing
	orderHandler.OrderPathPrefix = "/orders" //default is /order (signular)
	orderHandler.RegisterHandlers(a.Router)
	//Genre routing
	genreHandler.GenrePathPrefix = "/genres" //default is /genre (singular)
	genreHandler.RegisterHandlers(a.Router)
	//Core routing
	a.Router.HandleFunc("/healthcheck", a.healthCheck).Methods("GET", "POST")
	a.Router.HandleFunc("/healthcheck/panic", a.healthCheckPanic)
	a.Router.HandleFunc("/healthcheckadvanced", a.healthCheckAdvanced).Methods("GET", "POST")
	//Static routing
	//a.StaticRouter = a.Router.PathPrefix("/static").Subrouter()
	//a.StaticRouter.Handle("/static", http.StripPrefix("/static/", http.FileServer(http.Dir(".\\web\\static\\"))))
	//Root
	a.Router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			r.URL.Path = "/index.htm"
		}
		http.FileServer(http.Dir(".\\web\\static\\")).ServeHTTP(w, r)
	})
	//a.Router.PathPrefix("/").HandlerFunc(a.homeHandler)
	//a.Router.HandleFunc("/", a.homeHandler).Methods("GET")

	/*
		Panic/Internal error handler from pg. 59 of Web Programming with Go; Building and Scaling Interactive Web Applications with Go's Robust Ecosystem by Ian Taylor, 2023 GitforGits
		Learned that you can chain handlers in the Use()!
	*/
	a.Router.Use(func(handlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer InternalServerErrorHandler(w, r)
			handlr.ServeHTTP(w, r)
		})
	}, RateLimit, RequestThrottle)
}

/*
The healthCheck method returns a simple HTTP response and 200 (OK) status. If the Referer header
was provided, it will be reflected.
*/
func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	referrer := r.Referer()
	response := fmt.Sprintf("OK %v", referrer)
	w.Write([]byte(response))
}

/*
The healthChecAdvanced method checks more details of the server status and
returns a status code that reflects these facts.
*/
func (a *App) healthCheckAdvanced(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = err.Error()
	}
	pid := os.Getpid()
	stdResponse := fmt.Sprintf("Hostname: [%v], PID: [%v], Address: [%v]", hostname, pid, a.Configs.ServerAddress)

	//Check DB access
	if err := a.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("Database unreachable. %v", stdResponse)
		w.Write([]byte(response))
		log.Printf("/healthCheckAdvanced failure: %v [%v]", response, err)
		return
	}

	// Add more checks here

	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("OK. %v", stdResponse)
	w.Write([]byte(response))
}

/*
The healthChecPanic method induces a panic.
*/
func (a *App) healthCheckPanic(w http.ResponseWriter, r *http.Request) {
	panic("Panic! At the Disco")
}

/*
Handles internal server errors (HTTP Status 500) and prevents the server from dying.
*/
func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	//Anything outside this if below runs for every request to the server... don't put too much here.
	if r := recover(); r != nil {
		//Stuff in here will only run in the event of a panic somewhere else.
		log.Printf("Recovered from the following error: [%v]", r)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Well that's embarassing, it appears we've erred ourself. We'll go change and be right back."))
	}
}

/*
Enforces rate limits for all Handlers that .Use(RateLimit) it.
*/
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

/*
Throttles requests for all Handlers that .Use(RequestThrottle) it.
*/
func RequestThrottle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) //Introduces a 200ms delay for every request
		next.ServeHTTP(w, r)
	})
}

func (a *App) Run(addr string) {
	log.Printf("Launching on [%v]", addr)
	http.ListenAndServe(addr, a.Router)
}

/*
Loads HTML templates into cache
*/
func (a *App) loadTemplates() {
	templateLocation := "./web/templates/*.gohtml"
	templates, err := TemplateCache.ParseGlob(templateLocation)
	if err != nil {
		log.Fatalf("Could not load templates from [%v]. Error: %v", templateLocation, err)
	} else {
		log.Printf("Templates loaded: %v", len(templates.Templates()))
	}
	TemplateCache = templates
	//This is broken
	//bookHandler.LoadTemplates()
	bookHandler.SetTemplateCache(TemplateCache)
	genreHandler.SetTemplateCache(TemplateCache)
}

func (a *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	myData := HomeTemplate{
		PageTitle: "GitforGits Bookstore",
		PageBody:  "Hello GitforGits Bookstore!",
	}
	myTemplate := "index.gohtml"
	if err := TemplateCache.ExecuteTemplate(w, myTemplate, myData); err != nil {
		log.Printf("Could not execute %v, error: %v", myTemplate, err)
		http.Error(w, "Oops, we encountered an error. We're just going to blame this one on our editors.", http.StatusInternalServerError)
	}
	//log.Printf("Rendered template [%v]", myTemplate)
}

/*
Main function. This runs first.
*/
func main() {
	app := &App{}
	app.Initialize()
	defer app.DB.Close() //must happen in main because if it's done inside Initialize, the connection closes at the end of the function.
	app.loadTemplates()
	app.Run(app.Configs.ServerAddress)
}
