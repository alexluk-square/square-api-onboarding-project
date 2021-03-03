package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
)

func NewSquareOAuths(squareOAuth *oauth2.Config) *SquareOAuths {
	return &SquareOAuths{
		squareOAuth: squareOAuth,
	}
}

type SquareOAuths struct {
	squareOAuth *oauth2.Config
}

func (s *SquareOAuths) SquareConnect(w http.ResponseWriter, r *http.Request) {
	state := csrf.Token(r)
	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	url := s.squareOAuth.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *SquareOAuths) SquareCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	state := r.FormValue("state")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if stateCookie == nil || stateCookie.Value != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// clear cookie
	stateCookie.Value = ""
	stateCookie.Expires = time.Now()
	http.SetCookie(w, stateCookie)

	code := r.FormValue("code")
	token, err := s.squareOAuth.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenCookie := http.Cookie{
		Name:     "oauth_token",
		Value:    base64.StdEncoding.EncodeToString(data),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &tokenCookie)

	fmt.Fprintf(w, "%v", token)
}

func (s *SquareOAuths) ListCustomers(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("oauth_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if tokenCookie == nil {
		http.Error(w, "oauth_token cookie does not exist", http.StatusBadRequest)
		return
	}

	var token oauth2.Token
	data, err := base64.StdEncoding.DecodeString(tokenCookie.Value)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &token); err != nil {
		panic(err)
	}

	client := s.squareOAuth.Client(context.TODO(), &token)

	req, err := http.NewRequest(http.MethodGet,
		"https://connect.squareup.com/v2/customers",
		strings.NewReader("{\"path\": \"\"}"))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}
