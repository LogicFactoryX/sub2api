package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestMembershipRepositoryGrantOrExtendUsesExplicitPostgresTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, time.July, 16, 8, 30, 0, 0, time.UTC)
	expectedExpiry := now.Add(30 * 24 * time.Hour)
	mock.ExpectQuery(`(?s)INSERT INTO user_memberships.*\$2::timestamptz.*\$3::integer.*GREATEST\(user_memberships\.expires_at, \$2::timestamptz\).*\$3::integer.*RETURNING expires_at`).
		WithArgs(int64(12), now, 30, int64(34)).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expectedExpiry))

	repo := NewMembershipRepository(client)
	got, err := repo.GrantOrExtend(context.Background(), 12, 34, 30, now)
	require.NoError(t, err)
	require.Equal(t, expectedExpiry, got)
	require.NoError(t, mock.ExpectationsWereMet())
}
