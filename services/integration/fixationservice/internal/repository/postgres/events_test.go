package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/shared/pkg/authz"
	"Broker_backend/shared/pkg/dbtest"
	"Broker_backend/shared/pkg/requestctx"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryWritesAuditAndOutboxContract(t *testing.T) {
	instance := dbtest.Start(t)
	t.Cleanup(instance.Close)

	// Репозиторий в приложении работает под go_user. Проверяем теми же
	// правами, иначе тест под postgres пропустит ошибку в грантах или RLS.
	goUserDSN := instance.DSNForRole("go_user", "go_user", "integration,app,public")
	pool, err := pgxpool.New(context.Background(), goUserDSN)
	if err != nil {
		t.Fatalf("create go_user pool: %v", err)
	}
	t.Cleanup(pool.Close)

	repository := NewRepository(NewTxManager(pool))
	fixedAt := time.Date(2026, time.August, 20, 13, 0, 0, 0, time.UTC)
	fixation := entity.Fixation{
		FixationID: uuid.MustParse("018f47f2-7b86-7a03-b9ea-0242ac120002"),
		AgencyID:   uuid.MustParse("8e121fa4-635e-40ca-8391-412707401210"),
		FixBy:      uuid.MustParse("64f32c50-af66-4d09-8af3-3682833282f0"),
		FixFor:     uuid.MustParse("64f32c50-af66-4d09-8af3-3682833282f0"),
		ProjectID:  uuid.MustParse("459e2d4d-151e-4d0e-a46c-87a53eff7d1c"),
		PhoneHash:  "signed-phone",
		Status:     entity.StatusActive,
		FixedAt:    fixedAt,
		ExpiresAt:  fixedAt.Add(30 * 24 * time.Hour),
	}

	principal := authz.Principal{
		AgencyID: fixation.AgencyID,
		UserID:   fixation.FixBy,
		DeviceID: "repository-test",
		Role:     "agency_owner",
	}
	ctx := authz.WithPrincipal(context.Background(), principal)
	ctx = requestctx.WithRequestID(ctx, "request-events-test")
	ctx = requestctx.WithClientIP(ctx, "192.0.2.25")

	err = repository.Tx.Do(ctx, func(txCtx context.Context) error {
		if err := repository.InsertNewFixation(txCtx, fixation); err != nil {
			return err
		}
		if err := repository.InsertAudit(txCtx, fixation); err != nil {
			return err
		}

		return repository.InsertOutbox(txCtx, fixation)
	})
	if err != nil {
		t.Fatalf("write fixation transaction: %v", err)
	}

	var (
		entityType    string
		action        string
		actorUserID   uuid.UUID
		actorAgencyID uuid.UUID
		actorRole     string
		requestID     string
		clientIP      string
		auditState    []byte
	)
	err = pool.QueryRow(ctx, `
		select entity_type, action, actor_user_id, actor_agency_id,
		       actor_role, request_id, client_ip::text, state_after
		  from integration.audit_log
		 where entity_id = $1`, fixation.FixationID).Scan(
		&entityType,
		&action,
		&actorUserID,
		&actorAgencyID,
		&actorRole,
		&requestID,
		&clientIP,
		&auditState,
	)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}

	if entityType != fixationAggregateType || action != fixationCreatedAction {
		t.Fatalf("unexpected audit kind: entity_type=%q action=%q", entityType, action)
	}
	if actorUserID != principal.UserID || actorAgencyID != principal.AgencyID || actorRole != principal.Role {
		t.Fatalf("audit actor does not match principal")
	}
	if requestID != "request-events-test" {
		t.Fatalf("unexpected request_id: %q", requestID)
	}
	if clientIP != "192.0.2.25/32" {
		t.Fatalf("unexpected client_ip: %q", clientIP)
	}

	var (
		aggregateType string
		aggregateID   uuid.UUID
		eventType     string
		outboxPayload []byte
	)
	err = pool.QueryRow(ctx, `
		select aggregate_type, aggregate_id, event_type, payload
		  from integration.outbox
		 where aggregate_id = $1`, fixation.FixationID).Scan(
		&aggregateType,
		&aggregateID,
		&eventType,
		&outboxPayload,
	)
	if err != nil {
		t.Fatalf("read outbox: %v", err)
	}

	if aggregateType != fixationAggregateType || aggregateID != fixation.FixationID || eventType != fixationCreatedEventType {
		t.Fatalf(
			"unexpected outbox envelope: aggregate_type=%q aggregate_id=%s event_type=%q",
			aggregateType,
			aggregateID,
			eventType,
		)
	}

	assertFixationState(t, auditState, fixation)
	assertFixationState(t, outboxPayload, fixation)

	tag, err := pool.Exec(ctx, `
		update integration.audit_log
		   set action = 'removed'
		 where entity_id = $1`, fixation.FixationID)
	assertAuditMutationDenied(t, "update", tag.RowsAffected(), err)

	tag, err = pool.Exec(ctx, `
		delete from integration.audit_log
		 where entity_id = $1`, fixation.FixationID)
	assertAuditMutationDenied(t, "delete", tag.RowsAffected(), err)
}

func assertAuditMutationDenied(t *testing.T, operation string, rowsAffected int64, err error) {
	t.Helper()

	if err == nil {
		if rowsAffected != 0 {
			t.Fatalf("append-only audit allowed %s", operation)
		}
		return
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != pgerrcode.InsufficientPrivilege {
		t.Fatalf("attempt audit %s: %v", operation, err)
	}
}

func assertFixationState(t *testing.T, raw []byte, want entity.Fixation) {
	t.Helper()

	var got fixationState
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode fixation state: %v", err)
	}

	if got.FixationID != want.FixationID ||
		got.AgencyID != want.AgencyID ||
		got.FixBy != want.FixBy ||
		got.FixFor != want.FixFor ||
		got.ProjectID != want.ProjectID ||
		got.PhoneHash != want.PhoneHash ||
		got.Status != want.Status ||
		!got.FixedAt.Equal(want.FixedAt) ||
		!got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Fatalf("unexpected fixation state: %#v", got)
	}
}
