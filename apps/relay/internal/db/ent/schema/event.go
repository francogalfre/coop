package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Event struct {
	ent.Schema
}

func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.Int("seq").NonNegative(),
		field.Bytes("data").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", AgentSession.Type).Ref("events").Unique().Required(),
	}
}

func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("session").Fields("seq").Unique(),
	}
}
