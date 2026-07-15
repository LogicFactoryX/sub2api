package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserGroupEntitlement struct {
	ent.Schema
}

func (UserGroupEntitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_group_entitlements"}}
}

func (UserGroupEntitlement) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Int64("fallback_group_id"),
		field.Int64("redeem_code_id").Optional().Nillable().Unique(),
		field.Time("starts_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("status").MaxLen(20).Default("active"),
		field.Time("expired_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (UserGroupEntitlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "group_id", "expires_at"),
		index.Fields("status", "expires_at"),
	}
}
