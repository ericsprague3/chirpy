package error

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"

	internalError "github.com/ericsprague3/chirpy/internal/types"
)

func HandleErrorResponse(resp http.ResponseWriter, err error, logMessage, errorMessage string, errorStatus int) {
	log.Printf("%s: %s\n", logMessage, err)
	respBody := internalError.ErrorBody{
		Error: errorMessage,
	}

	// Postgres Checks
	// Check for invalid uuid insert error. Code: 23503
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.UniqueViolation {
		respBody.Error = "Invalid UUID, Unique Error Constraint"

	}

	data, marErr := json.Marshal(respBody)
	if marErr != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		resp.WriteHeader(500)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(errorStatus)
	resp.Write(data)
}
