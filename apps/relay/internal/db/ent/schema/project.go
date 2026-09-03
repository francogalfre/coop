package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Project struct {
	ent.Schema
}

func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").NotEmpty().Unique(),
		field.String("created_by").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.String("context_text").Default(""),
		field.Int("context_version").Default(0).NonNegative(),
		field.String("context_updated_by").Default(""),
		field.Time("context_updated_at").Optional().Nillable(),
	}
}

func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ProjectMember.Type),
		edge.To("invites", ProjectInvite.Type),
		edge.To("sessions", AgentSession.Type),
		edge.To("agents", Agent.Type),
		edge.To("notes", ProjectNote.Type),
	}
}
