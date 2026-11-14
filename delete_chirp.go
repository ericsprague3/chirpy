package main

import (
	"net/http"

	"github.com/ericsprague3/chirpy/internal/database"
	internalError "github.com/ericsprague3/chirpy/internal/error"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleDeleteChirp(resp http.ResponseWriter, req *http.Request) {
	requestCtx := req.Context()

	chirpId, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		logMessage := "Error parsing chirp ID from path"
		errorMessage := "Invalid chirp ID, must be UUID"
		errorStatus := 400
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	userID, ok := requestCtx.Value(userIDKey).(uuid.UUID)
	if !ok {
		http.Error(resp, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbResp, err := cfg.dbQueries.GetChirpById(requestCtx, chirpId)
	if err != nil {
		logMessage := "Error fetching chirp for delete chirp"
		errorMessage := "Chirp not found"
		errorStatus := 404
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	if dbResp.UserID != userID {
		logMessage := "Unauthorized user attempting to delete chirp"
		errorMessage := "Unauthorized"
		errorStatus := 403
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	err = cfg.dbQueries.DeleteChirpByIdAndUserId(requestCtx, database.DeleteChirpByIdAndUserIdParams{
		ID:     chirpId,
		UserID: userID,
	})
	if err != nil {
		logMessage := "Error deleting chirp"
		errorMessage := "Something went wrong"
		errorStatus := 500
		internalError.HandleErrorResponse(resp, err, logMessage, errorMessage, errorStatus)
		return
	}

	resp.WriteHeader(204)
}
