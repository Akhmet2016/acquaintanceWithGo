package app

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"lens/utils"
	"net/http"
	"os"
	models2 "second-rest-api/models"
	utils2 "second-rest-api/utils"
	"strings"
)

var JwtAuthentication = func(next http.Handler) http.Handler {
	return http.HamdlerFunc(func(w http.ResponseWrite, r *http.Request) {
		notAuth := []string{"/api/user/new", "/api/user/login"}
		requestPath := r.URL.Path
		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string] interface{})
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" {
			response = utils2.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils2.Respond(w, response)
			return
		}

		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			response = utils2.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils2.Respond(w, response)
			return
		}

		tokenPart := splitted[1]
		tk := &models2.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil {
			response = utils2.Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils2.Respond(w, response)
			return
		}

		if !token.Valid {
			response = utils2.Message(false, "Token is not valid")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils2.Respond(w, response)
			return
		}

		fmt.Printf("User %", tk.Username)
		ctx := context.WithValue(r.Context(), "user", tk.UserId)
		r = r.WithContext(ctx)
		next.ServerHTTP(w, r)
	})
}
