package cors

import "net/http"

func CORSHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		nextHandler.ServeHTTP(w, r)
	})
}
