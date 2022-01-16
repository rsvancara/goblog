package middleware

import (
	"goblog/internal/sessionmanager"
	"net/http"
)

// AuthHandlerMiddleware authorize user
func (mw *MiddleWareContext) AuthHandlerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var sess sessionmanager.Session
		err := sess.Session(*mw.cache, mw.hConfig.RedisDB, r, w)

		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if !sess.User.IsAuth {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		h.ServeHTTP(w, r)
	})
}
