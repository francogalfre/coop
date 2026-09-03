package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProjectNote struct {
	ent.Schema
}

func (ProjectNote) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Immutable(),
		field.String("author_id").NotEmpty(),
		field.String("author_display_name").Default(""),
		field.String("author_avatar_url").Default(""),
		field.String("source").Default("human"),
		field.String("session_id").Optional(),
		field.String("text").NotEmpty(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (ProjectNote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("notes").Unique().Required(),
	}
}

func (ProjectNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").Fields("created_at"),
	}
}
