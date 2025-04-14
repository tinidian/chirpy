package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tinidian/chirpy/internal/auth"
	"github.com/tinidian/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type parameters struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type response struct {
	User
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hash, err := auth.HashedPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hash})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	generateUserResponse(w, http.StatusCreated, user)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect eamil or password", err)
		return
	}

	auth := auth.CheckPasswordHash(user.HashedPassword, params.Password)
	if auth != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect eamil or password", err)
		return
	}
	generateUserResponse(w, http.StatusOK, user)
}

func generateUserResponse(w http.ResponseWriter, code int, user database.User) {
	respondWithJSON(w, code, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})

}
