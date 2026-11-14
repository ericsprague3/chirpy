package main

import (
	"encoding/json"
	"net/http"

	"github.com/ericsprague3/chirpy/internal/auth"
	internalError "github.com/ericsprague3/chirpy/internal/error"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleWebHookUserUpgraded(resp http.ResponseWriter, req *http.Request) {
	requestCtx := req.Context()

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	// Confirm valid ApiKey in Authorization Header
	key, err := auth.GetAPIKey(req.Header)
	if err != nil || key != cfg.polkaKey {
		resp.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	// Error on decoding incoming request body
	if err != nil {
		logMessage := "Error decoding parameters"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	if params.Event != "user.upgraded" {
		resp.WriteHeader(204)
		return
	}

	// Confirm user exists
	dbResp, err := cfg.dbQueries.GetUserById(requestCtx, params.Data.UserId)
	if err != nil {
		logMessage := "Error fetching user by id in upgrade user"
		errorMessage := "Something went wrong"
		errorStatus := 404
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	err = cfg.dbQueries.UpgradeUser(requestCtx, dbResp.ID)
	if err != nil {
		logMessage := "Error upgrading user"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	resp.WriteHeader(204)
}
