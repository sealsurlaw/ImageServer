package middle

import (
	"fmt"
	"net/http"
	"strings"
)

func LogRoutes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, strings.Split(r.URL.String(), "?")[0])
		next.ServeHTTP(w, r)
	})
}
