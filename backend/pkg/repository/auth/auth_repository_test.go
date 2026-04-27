package auth

import (
	"context"
	"regexp"
	"testing"

	"flowforge-automation-backend/pkg/model/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockedRepo(t *testing.T) (Repository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{})
	assert.NoError(t, err)

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return NewAuthRepository(gormDB), mock, cleanup
}

func TestGetUserByEmailFound(t *testing.T) {
	repo, mock, cleanup := newMockedRepo(t)
	defer cleanup()

	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "email", "password_hash", "role", "is_active"}).
		AddRow(userID, uuid.New(), "admin@acme.com", "hash", "admin", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin@acme.com", 1).
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail(context.Background(), "admin@acme.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin@acme.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmailNotFound(t *testing.T) {
	repo, mock, cleanup := newMockedRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("missing@acme.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetUserByEmail(context.Background(), "missing@acme.com")

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTenant(t *testing.T) {
	repo, mock, cleanup := newMockedRepo(t)
	defer cleanup()

	tenant := &domain.Tenant{BaseModel: domain.BaseModel{ID: uuid.New()}, Name: "Acme", Slug: "acme"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tenants" .* VALUES .* RETURNING`).
		WithArgs(nil, nil, "Acme", "acme", tenant.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(tenant.ID, nil, nil))
	mock.ExpectCommit()

	err := repo.CreateTenant(context.Background(), tenant)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
