package middleware

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
	"goblog/internal/session"
	"goblog/internal/util"
)

// SessionMiddleware manage session objects
func (mw *MiddleWareContext) SessionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sess session.Session
		err := sess.Session(r, w)
		if err != nil {
			log.Error().Err(err).Str("service", "session").Msg("Error creating session in session middleware")
		}

		var ctxKey util.CtxKey
		ctxKey = "session"
		ctx := context.WithValue(r.Context(), ctxKey, sess)

		//fmt.Printf("session token created %s", sess.SessionToken)

		h.ServeHTTP(w, r.WithContext(ctx))

	})
}
