package cors

import "net/http"

const BaseFrontendPath string = "http://localhost:5173"

func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", BaseFrontendPath)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
