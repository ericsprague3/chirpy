package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	auth "github.com/ericsprague3/chirpy/internal/auth"
	"github.com/ericsprague3/chirpy/internal/database"
	internalError "github.com/ericsprague3/chirpy/internal/error"
)

func (cfg *apiConfig) HandleLoginUser(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseBody struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	requestCtx := req.Context()

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	// Error on decoding incoming request body
	if err != nil {
		logMessage := "Error decoding parameters"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	// Fetch User
	user, err := cfg.dbQueries.GetUserByEmail(requestCtx, params.Email)
	if err != nil {
		logMessage := "User not found with provided email"
		errorMessage := "Unauthorized"
		errorStatus := 401
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	// Hash Password
	valid, err := auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil || !valid {
		logMessage := "Invalid Password provided"
		errorMessage := "Unauthorized"
		errorStatus := 401
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	// Create JWT
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Duration(DEFAULT_EXPIRES)*time.Second)
	if err != nil {
		logMessage := "Error when creating JWT"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	// Create Refresh Token
	refreshToken := auth.MakeRefreshToken()

	// Insert Refresh Token
	_, err = cfg.dbQueries.CreateRefreshToken(requestCtx, database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
	})

	if err != nil {
		logMessage := "Error when inserting Refresh Token"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	respBody := responseBody{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	}

	// Valid decode
	data, marErr := json.Marshal(respBody)
	if marErr != nil {
		resp.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s\n", err)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(200)
	resp.Write(data)
}
