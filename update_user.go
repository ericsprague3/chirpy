package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ericsprague3/chirpy/internal/auth"
	"github.com/ericsprague3/chirpy/internal/database"
	internalError "github.com/ericsprague3/chirpy/internal/error"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleUpdateUser(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type errorBody struct {
		Error string `json:"error"`
	}

	type responseBody struct {
		ID          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	requestCtx := req.Context()

	// 1. Extract userID from context value key userIDKey
	userID, ok := requestCtx.Value(userIDKey).(uuid.UUID)
	if !ok {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Hash Password
	password, err := auth.HashPassword(params.Password)
	if err != nil {
		logMessage := "Error hashing password"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	dbResp, err := cfg.dbQueries.UpdateUser(requestCtx, database.UpdateUserParams{
		Email:    params.Email,
		Password: password,
		ID:       userID,
	})
	if err != nil {
		logMessage := "Error updating user"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	respBody := responseBody{
		ID:          dbResp.ID.String(),
		CreatedAt:   dbResp.CreatedAt,
		UpdatedAt:   dbResp.UpdatedAt,
		Email:       dbResp.Email,
		IsChirpyRed: dbResp.IsChirpyRed,
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
