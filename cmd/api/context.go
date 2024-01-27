package main

import (
	"context"
	"net/http"

	"movies.samkha.net/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

// returns copy of request with provided User struct
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)

	if !ok {
		panic("missing user value in request context")
	}

	return user
}
