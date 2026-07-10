package fixationcustomer

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/infra/repositories/postgres"
	"context"
	"time"
)

type Repository struct {
	pg *postgres.Postgres
}

func NewRepository(pg *postgres.Postgres) *Repository {
	return &Repository{
		pg: pg,
	}
}

func (r *Repository) SaveFixationCustomer(ctx context.Context, brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID, now *time.Time) error { //как сделать чтобы юзкейс видел вызов этого метода? как принято делать?
	const op = "postgres.fixationcustomer.SaveFixationCustomer" //есть ли смысл в моем случае писать op? стоит убрать его из мапера и в целом не использовать их?
	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	query := `UPDATE customers SET manager_id = $1 WHERE customer_id = $2`

	_, err := r.pg.DB().Exec(ctx, query, managerID, customerID) //не понимаю почему r.pg.DB() а не просто r.pg, r.pg же и есть *postgres.Postgres, почему вот так у меня и правильно ли так делать и почему?
	if err != nil {
		return postgres.MapError(op, err)
	}

	return nil
}
