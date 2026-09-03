package httpapi

import "github.com/francogalfre/coop/apps/relay/internal/auth"

func actorJSON(a auth.Actor) map[string]string {
	fields := map[string]string{"id": a.UserID, "display_name": a.DisplayName}
	if a.AvatarURL != "" {
		fields["avatar_url"] = a.AvatarURL
	}

	return fields
}
