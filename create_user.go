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

func (cfg *apiConfig) HandleCreateUser(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseBody struct {
		ID          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
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

	// Hash Password
	password, err := auth.HashPassword(params.Password)
	if err != nil {
		logMessage := "Error hashing password"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	// Create User
	dbResp, err := cfg.dbQueries.CreateUser(requestCtx, database.CreateUserParams{
		Email: params.Email, Password: password,
	})
	if err != nil {
		logMessage := "Error decoding parameters"
		errorMessage := "Failed to create user"
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
	resp.WriteHeader(201)
	resp.Write(data)
}
