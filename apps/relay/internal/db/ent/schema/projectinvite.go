package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ProjectInvite struct {
	ent.Schema
}

func (ProjectInvite) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("token_hash").NotEmpty().Unique(),
		field.String("created_by").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional().Nillable(),
	}
}

func (ProjectInvite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("invites").Unique().Required(),
	}
}
