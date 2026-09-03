package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type AgentSession struct {
	ent.Schema
}

func (AgentSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Immutable(),
		field.String("owner_id").NotEmpty(),
		field.String("repo").Default(""),
		field.String("cwd").Default(""),
		field.String("harness").Default(""),
		field.String("status").Default("live"),
		field.String("mode").Default("auto"),
		field.Int("next_seq").Default(0).NonNegative(),
		field.Time("started_at"),
		field.Time("ended_at").Optional().Nillable(),
	}
}

func (AgentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("sessions").Unique().Required(),
		edge.From("agent", Agent.Type).Ref("sessions").Unique(),
		edge.To("events", Event.Type),
		edge.To("steer_requests", SteerRequest.Type),
		edge.To("takeover", Takeover.Type).Unique(),
	}
}
