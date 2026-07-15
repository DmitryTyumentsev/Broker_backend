package fixationcustomer

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/infra/repositories/postgres"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	pg *postgres.Postgres
}

func NewRepository(pg *postgres.Postgres) *Repository {
	return &Repository{
		pg: pg,
	}
}

func (r *Repository) SaveFixationCustomer(ctx context.Context, uuid string, expiresAt *time.Time, fixedAt *time.Time, status string, brokerID entity.BrokerID, fixedBy entity.fixedBy, managerID entity.ManagerID, customerID entity.CustomerID) error { //как сделать чтобы юзкейс видел вызов этого метода? как принято делать?
	const op = "postgres.fixationcustomer.SaveFixationCustomer"
	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	query := `INSERT INTO customers(uuid, expires_at, fixed_at, status, broker_id, fixed_by, manager_id, customer_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pg.DB().Exec(ctx, query, uuid, expiresAt, fixedAt, status, brokerID, fixedBy, managerID, customerID)
	if err != nil {
		return postgres.MapError(op, err)
	}

	return nil
}

//func (r *Repository) UpdateFixationCustomer(ctx context.Context, brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID, now *time.Time) error { //как сделать чтобы юзкейс видел вызов этого метода? как принято делать?
//	const op = "postgres.fixationcustomer.UpdateFixationCustomer" //есть ли смысл в моем случае писать op? стоит убрать его из мапера и в целом не использовать их?
//	ctx, cancel := r.pg.WriteWithTimeout(ctx)
//	defer cancel()
//
//	query := `UPDATE customers SET broker_id = $1 AND manager_id = $2 AND updated_at = $3 WHERE customer_id = $4`
//
//	tag, err := r.pg.DB().Exec(ctx, query, brokerID, managerID, now, customerID) //не понимаю почему r.pg.DB() а не просто r.pg, r.pg же и есть *postgres.Postgres, почему вот так у меня и правильно ли так делать и почему?
//	if tag.RowsAffected() == 0 || err != nil {
//		return postgres.MapError(op, err)
//	}
//
//	return nil
//}
