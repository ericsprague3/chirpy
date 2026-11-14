package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) HandleMetrics(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(200)
	body := fmt.Sprintf(`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`, cfg.fileserverHits.Load())
	fmt.Fprint(resp, body)
}
