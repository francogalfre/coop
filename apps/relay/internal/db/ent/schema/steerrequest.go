package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type SteerRequest struct {
	ent.Schema
}

func (SteerRequest) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Immutable(),
		field.String("actor_id").NotEmpty(),
		field.String("actor_display_name").Default(""),
		field.String("actor_avatar_url").Default(""),
		field.String("text").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (SteerRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", AgentSession.Type).Ref("steer_requests").Unique().Required(),
	}
}
