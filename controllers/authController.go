package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/atinyakov/go_final_project/models"
)

type AuthController struct {
	jwtc *JwtController
}

func NewAuthController(jwtc *JwtController) *AuthController {
	return &AuthController{
		jwtc: jwtc,
	}
}

func (a *AuthController) HandleAuth(w http.ResponseWriter, req *http.Request) {
	pass := os.Getenv("TODO_PASSWORD")
	// test pass
	if pass == "" {
		pass = "321"
	}

	fmt.Println("got pass", pass)

	var authRequest models.Auth
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &authRequest); err != nil {
		http.Error(w, "Error: Cannot deserialize JSON", http.StatusInternalServerError)
		return
	}

	requestPassword := authRequest.Password

	if pass != requestPassword {
		errorResponce := &models.TaskResponceError{Error: "Incorrect password"}

		errorJSON, jsonErr := json.Marshal(errorResponce)
		if jsonErr != nil {
			http.Error(w, "Failed to encode error", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(errorJSON), http.StatusInternalServerError)
		return
	} else {

		token := a.jwtc.SignToken()

		cookie := &http.Cookie{
			Name:     "auth_token",                  // Name of the cookie
			Value:    token,                         // The signed JWT token
			HttpOnly: true,                          // Prevents JavaScript access to the cookie
			Path:     "/",                           // The path where the cookie is valid
			Expires:  time.Now().Add(8 * time.Hour), // Cookie expiration time should match JWT expiration (8 hours)
		}

		http.SetCookie(w, cookie)
		response := map[string]string{
			"token": token,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode token response", http.StatusInternalServerError)
			return
		}
	}

}

func (a *AuthController) Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		// test pass
		if pass == "" {
			pass = "321"
		}

		fmt.Println("got pass", pass)

		if len(pass) > 0 {
			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}

			valid := a.jwtc.Validate(jwt)

			if !valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
