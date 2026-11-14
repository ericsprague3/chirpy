package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ericsprague3/chirpy/internal/auth"
	internalError "github.com/ericsprague3/chirpy/internal/error"
)

func (cfg *apiConfig) HandleRefresh(resp http.ResponseWriter, req *http.Request) {
	type responseBody struct {
		Token string `json:"token"`
	}

	requestCtx := req.Context()

	// 1. Extract Bearer token from Authentication Header
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Lookup refresh token, if 1) non-existent; 2) expired; or 3) revoked
	// throw error
	dbResp, err := cfg.dbQueries.GetRefreshToken(requestCtx, refreshToken)
	if err != nil || time.Now().After(dbResp.ExpiresAt) || dbResp.RevokedAt.Valid {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Create JWT
	token, err := auth.MakeJWT(dbResp.UserID, cfg.tokenSecret, time.Duration(DEFAULT_EXPIRES)*time.Second)
	if err != nil {
		logMessage := "Error creating JWT"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	respBody := responseBody{
		Token: token,
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
