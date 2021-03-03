package main

import (
	"square-onboarding/controllers"
	"square-onboarding/views"
	"net/http"

	"square-onboarding/rand"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

var (
	homeView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func buildSquareOAuthController() *controllers.SquareOAuths {
	squareOAuth := &oauth2.Config{
		ClientID:     "**REDACTED**",
		ClientSecret: "**REDACTED**",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://connect.squareup.com/oauth2/authorize",
			TokenURL: "https://connect.squareup.com/oauth2/token",
		},
		RedirectURL: "http://localhost:3000/oauth/square/callback",
		Scopes:      []string{"CUSTOMERS_READ", "CUSTOMERS_WRITE"},
	}
	return controllers.NewSquareOAuths(squareOAuth)
}

func main() {
	r := mux.NewRouter()
	squareOAuthsC := buildSquareOAuthController()

	homeView = views.NewView("bootstrap", "views/home.gohtml")

	b, err := rand.Bytes(32)
	must(err)
	csrfMw := csrf.Protect(b)

	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/oauth/square/connect", squareOAuthsC.SquareConnect).Methods("GET")
	r.HandleFunc("/oauth/square/callback", squareOAuthsC.SquareCallback).Methods("GET")
	r.HandleFunc("/customers", squareOAuthsC.ListCustomers).Methods("GET")
	http.ListenAndServe(":3000", csrfMw(r))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
