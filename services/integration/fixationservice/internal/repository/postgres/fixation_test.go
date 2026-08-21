package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/shared/pkg/dbtest"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryFixationMethods(t *testing.T) {
	instance := dbtest.Start(t)
	t.Cleanup(instance.Close)

	ctx := context.Background()
	agencyID := uuid.New()
	otherAgencyID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	seedRepositoryReferences(t, ctx, instance, agencyID, otherAgencyID, projectID, userID)

	goUserDSN := instance.DSNForRole("go_user", "go_user", "integration,app,public")
	pool, err := pgxpool.New(ctx, goUserDSN)
	if err != nil {
		t.Fatalf("create go_user pool: %v", err)
	}
	t.Cleanup(pool.Close)

	repository := NewRepository(NewTxManager(pool))

	status, err := repository.StatusByProjectID(ctx, projectID)
	if err != nil || status == "" {
		t.Fatalf("existing project: status=%v err=%v", status, err)
	}
	status, err = repository.StatusByProjectID(ctx, uuid.New())
	if err != nil || status == "" {
		t.Fatalf("missing project: status=%v err=%v", status, err)
	}

	belongs, err := repository.IsUserIDInAgencyID(ctx, agencyID, userID)
	if err != nil || !belongs {
		t.Fatalf("user in agency: belongs=%v err=%v", belongs, err)
	}
	belongs, err = repository.IsUserIDInAgencyID(ctx, otherAgencyID, userID)
	if err != nil || belongs {
		t.Fatalf("user in another agency: belongs=%v err=%v", belongs, err)
	}

	current, err := repository.FixationCurrent(ctx, "repository-phone", projectID)
	if err != nil {
		t.Fatalf("empty current fixation: %v", err)
	}
	if current.Status != entity.StatusNoRows {
		t.Fatalf("empty current status: got %q", current.Status)
	}

	fixedAt := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
	first := entity.Fixation{
		FixationID: uuid.New(),
		AgencyID:   agencyID,
		FixBy:      userID,
		FixFor:     userID,
		ProjectID:  projectID,
		PhoneHash:  "repository-phone",
		Status:     entity.StatusActive,
		FixedAt:    fixedAt,
		ExpiresAt:  fixedAt.Add(30 * 24 * time.Hour),
	}
	if err := repository.InsertNewFixation(ctx, first); err != nil {
		t.Fatalf("insert fixation: %v", err)
	}

	current, err = repository.FixationCurrent(ctx, first.PhoneHash, first.ProjectID)
	if err != nil {
		t.Fatalf("current fixation: %v", err)
	}
	if current.FixationID != first.FixationID || current.Status != entity.StatusActive {
		t.Fatalf("unexpected current fixation: %#v", current)
	}

	if err := repository.UpdateFixationStatusExpired(ctx, entity.StatusExpired, first.FixationID); err != nil {
		t.Fatalf("expire fixation: %v", err)
	}
	current, err = repository.FixationCurrent(ctx, first.PhoneHash, first.ProjectID)
	if err != nil {
		t.Fatalf("current expired fixation: %v", err)
	}
	if current.Status != entity.StatusExpired {
		t.Fatalf("status after expiration: got %q", current.Status)
	}

	if err := repository.UpdateFixationStatusExpired(ctx, entity.StatusActive, first.FixationID); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("invalid target status: got %v", err)
	}
	if err := repository.UpdateFixationStatusExpired(ctx, entity.StatusExpired, uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing fixation update: got %v", err)
	}
}

func seedRepositoryReferences(
	t *testing.T,
	ctx context.Context,
	instance *dbtest.Instance,
	agencyID uuid.UUID,
	otherAgencyID uuid.UUID,
	projectID uuid.UUID,
	userID uuid.UUID,
) {
	t.Helper()

	_, err := instance.Pool.Exec(ctx, `
		insert into app.agencies(id, name, status)
		values ($1, 'repository agency', 'active'),
		       ($2, 'other repository agency', 'active')`, agencyID, otherAgencyID)
	if err != nil {
		t.Fatalf("seed agencies: %v", err)
	}

	_, err = instance.Pool.Exec(ctx, `
		insert into app.projects(id, name, status)
		values ($1, 'repository project', 'active')`, projectID)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	_, err = instance.Pool.Exec(ctx, `
		insert into app.users(
			id, email, user_role, password_hash, last_name, first_name, agency_id
		) values ($1, $2, 'broker_team_member', 'x', 'Тестов', 'Тест', $3)`,
		userID, userID.String()+"@repository.test", agencyID)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
}
