package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ericsprague3/chirpy/internal/database"
	internalError "github.com/ericsprague3/chirpy/internal/error"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleGetChirps(resp http.ResponseWriter, req *http.Request) {
	type errorBody struct {
		Error string `json:"error"`
	}

	// Request Context
	requestContext := req.Context()

	// Blank dbResp
	var dbResp []database.Chirp

	// Sort direction, default to asc if not provided
	sort := "asc"
	sortQueryParam := req.URL.Query().Get("sort")
	if sortQueryParam != "" && (sortQueryParam == "asc" || sortQueryParam == "desc") {
		sort = sortQueryParam
	}

	// Author Id provided?
	authorId := req.URL.Query().Get("author_id")
	if authorId != "" {
		// Query using additional Where clause with user_id = author_id
		userId, err := uuid.Parse(req.URL.Query().Get("author_id"))
		if err != nil {
			logMessage := "Invalid author id, not UUID formatted"
			errorMessage := "Invalid author id"
			errorStatus := 400
			internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
			return
		}
		if sort == "asc" {
			dbResp, err = cfg.dbQueries.GetChirpsByUserIdAsc(requestContext, userId)
		} else {
			dbResp, err = cfg.dbQueries.GetChirpsByUserIdDesc(requestContext, userId)
		}
		if err != nil {
			logMessage := fmt.Sprintf("Error fetching chirps: %s\n", err)
			errorMessage := "Error fetching chirps"
			errorStatus := 500
			internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
			return
		}

	} else {
		var err error
		if sort == "asc" {
			dbResp, err = cfg.dbQueries.GetChirpsAsc(requestContext)
		} else {
			dbResp, err = cfg.dbQueries.GetChirpsDesc(requestContext)
		}

		if err != nil {
			logMessage := fmt.Sprintf("Error fetching chirps: %s\n", err)
			errorMessage := "Error fetching chirps"
			errorStatus := 500
			internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
			return
		}
	}

	// Valid decode
	data, marErr := json.Marshal(dbResp)
	if marErr != nil {
		resp.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s\n", marErr)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(200)
	resp.Write(data)
}

func (cfg *apiConfig) HandleGetChirpById(resp http.ResponseWriter, req *http.Request) {
	type errorBody struct {
		Error string `json:"error"`
	}

	// Request Context
	requestContext := req.Context()
	userId, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing chirp id: %s\n", err)
		respBody := errorBody{
			Error: fmt.Sprintf("Error parsing chirp id %s", userId),
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

	dbResp, err := cfg.dbQueries.GetChirpById(requestContext, userId)
	if err != nil {
		log.Printf("Error fetching chirp by id: %s\n", err)
		if errors.Is(err, sql.ErrNoRows) {
			resp.Header().Set("Content-Type", "application/json")
			resp.WriteHeader(404)
			return
		}

		respBody := errorBody{
			Error: fmt.Sprintf("Error fetching chirp %s", userId),
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

	// Valid decode
	data, marErr := json.Marshal(dbResp)
	if marErr != nil {
		resp.WriteHeader(500)
		log.Printf("Error marshalling JSON: %s\n", err)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(200)
	resp.Write(data)
}
