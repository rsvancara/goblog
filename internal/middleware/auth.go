package middleware

import (
	"net/http"

	"goblog/internal/session"
)

// AuthHandlerMiddleware authorize user
func (mw *MiddleWareContext) AuthHandlerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var sess session.Session

		err := sess.Session(r, w)

		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if sess.User.IsAuth == false {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		h.ServeHTTP(w, r)
	})
}
