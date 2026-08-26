package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CliCredential struct {
	ent.Schema
}

func (CliCredential) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty(),
		field.String("display_name").Default(""),
		field.Bytes("token_hash").NotEmpty().Unique(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("expires_at").Optional().Nillable(),
	}
}
