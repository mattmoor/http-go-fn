package foo

import "net/http"

// Wrong argument order.
func Handler(r *http.Request, w http.ResponseWriter) {
}
