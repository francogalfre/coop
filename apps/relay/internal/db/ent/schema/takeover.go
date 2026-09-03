package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// The id is the owning session's id: a takeover is at most one live row per session.
type Takeover struct {
	ent.Schema
}

func (Takeover) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Immutable(),
		field.String("actor_id").NotEmpty(),
		field.String("actor_display_name").Default(""),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (Takeover) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", AgentSession.Type).Ref("takeover").Unique().Required(),
	}
}
