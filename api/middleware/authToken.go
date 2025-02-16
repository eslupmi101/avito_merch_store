package authTokenMiddleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

func Authorization(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			values, ok := ctx.Value(config.AuthMiddlewareValuesKey).(map[string]interface{})
			if !ok {
				values = make(map[string]interface{})
			}
			values["isAuthorized"] = false

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ctx = context.WithValue(ctx, config.AuthMiddlewareValuesKey, values)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				slog.Debug("authHeader")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.Split(authHeader, " ")[1]

			valid, err := utility.IsAuthorized(token, secret)
			if err != nil || !valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			values["isAuthorized"] = true

			userID, err := utility.ExtractIDFromToken(token, secret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			values["userID"] = userID

			ctx = context.WithValue(ctx, config.AuthMiddlewareValuesKey, values)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
