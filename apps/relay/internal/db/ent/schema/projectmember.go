package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProjectMember struct {
	ent.Schema
}

func (ProjectMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty(),
		field.String("role").NotEmpty(),
		field.Time("joined_at").Immutable().Default(time.Now),
	}
}

func (ProjectMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("members").Unique().Required(),
	}
}

func (ProjectMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").Fields("user_id").Unique(),
	}
}
