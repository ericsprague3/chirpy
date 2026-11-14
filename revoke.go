package main

import (
	"net/http"

	"github.com/ericsprague3/chirpy/internal/auth"
	internalError "github.com/ericsprague3/chirpy/internal/error"
)

func (cfg *apiConfig) HandleRevoke(resp http.ResponseWriter, req *http.Request) {
	requestCtx := req.Context()

	// 1. Extract Bearer token from Authentication Header
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Revoke refresh token
	err = cfg.dbQueries.RevokeRefreshToken(requestCtx, refreshToken)
	if err != nil {
		logMessage := "Error revoking refresh token"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	resp.WriteHeader(204)
}
