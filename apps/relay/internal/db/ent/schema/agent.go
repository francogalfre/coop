package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Agent struct {
	ent.Schema
}

func (Agent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Immutable(),
		field.String("name").NotEmpty(),
		field.String("display_name").Default(""),
		field.String("created_by").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (Agent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("agents").Unique().Required(),
		edge.To("sessions", AgentSession.Type),
	}
}

func (Agent) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").Fields("name").Unique(),
	}
}
