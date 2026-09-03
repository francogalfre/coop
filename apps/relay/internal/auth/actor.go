package auth

import "context"

type Actor struct {
	UserID      string
	DisplayName string
	AvatarURL   string
}

type contextKey struct{}

var actorContextKey = contextKey{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey, actor)
}

func FromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorContextKey).(Actor)
	return actor, ok
}
