package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/ericsprague3/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

func (cfg *apiConfig) HandleCreateChirp(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorBody struct {
		Error string `json:"error"`
	}

	type responseBody struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    string    `json:"user_id"`
	}

	// Request Context
	requestCtx := req.Context()

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
		log.Printf("Error decoding parameters: %s\n", err)
		respBody := errorBody{
			Error: "Something went wrong",
		}
		data, marErr := json.Marshal(respBody)
		if marErr != nil {
			log.Printf("Error marshalling JSON: %s\n", err)
			resp.WriteHeader(500)
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(500)
		resp.Write(data)
		return
	}

	// Check Length is less than 140 characters
	characterCount := utf8.RuneCountInString(params.Body)
	if characterCount > 140 {
		log.Printf("Validation Error, length of body is %d\n", characterCount)
		respBody := errorBody{
			Error: "Chirp is too long",
		}
		data, marErr := json.Marshal(respBody)
		if marErr != nil {
			log.Printf("Error marshalling JSON: %s\n", err)
			resp.WriteHeader(500)
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(400)
		resp.Write(data)
		return
	}

	// Check Profanity
	cleanedBody := CheckProfanity(params.Body)

	// Create Chirp
	dbResp, err := cfg.dbQueries.CreateChip(requestCtx, database.CreateChipParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Error Inserting Chirp: %s\n", err)
		respBody := errorBody{
			Error: "Unable to insert chirp",
		}

		// Check for invalid uuid insert error. Code: 23503
		if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.ForeignKeyViolation {
			respBody.Error = "Invalid UUID, FK Error"

		}

		data, marErr := json.Marshal(respBody)
		if marErr != nil {
			log.Printf("Error marshalling JSON: %s\n", err)
			resp.WriteHeader(500)
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(500)
		resp.Write(data)
		return
	}

	// Response Body
	respBody := responseBody{
		ID:        dbResp.ID.String(),
		CreatedAt: dbResp.CreatedAt,
		UpdatedAt: dbResp.UpdatedAt,
		Body:      dbResp.Body,
		UserId:    dbResp.UserID.String(),
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
