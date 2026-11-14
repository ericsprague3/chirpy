package main

import (
	"errors"
	"net/http"
	"os"
)

func (cfg *apiConfig) HandleDeleteUsers(resp http.ResponseWriter, req *http.Request) error {
	requestCtx := req.Context()

	// Check PLATFORM env value
	platformEnv := os.Getenv("PLATFORM")
	if platformEnv != "dev" {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(403)
		return errors.New("Missing PLATFORM env variable - 403 Forbidden")
	}

	// Delete Users
	err := cfg.dbQueries.DeleteUsers(requestCtx)
	if err != nil {
		return err
	}
	return nil
}
