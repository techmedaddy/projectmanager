package auth

import "context"

type contextKey string

const currentUserContextKey contextKey = "current_user"

// CurrentUser contains the authenticated user details exposed to handlers.
type CurrentUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// WithCurrentUser stores the authenticated user in request context.
func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserContextKey, user)
}

// CurrentUserFromContext fetches the authenticated user from request context.
func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	user, ok := ctx.Value(currentUserContextKey).(CurrentUser)
	return user, ok
}
