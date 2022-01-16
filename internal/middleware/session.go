package middleware

import (
	"context"
	"net/http"

	"goblog/internal/sessionmanager"
	"goblog/internal/util"

	"github.com/rs/zerolog/log"
)

// SessionMiddleware manage session objects
func (mw *MiddleWareContext) SessionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sess sessionmanager.Session
		err := sess.Session(*mw.cache, mw.hConfig.RedisDB, r, w)
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
