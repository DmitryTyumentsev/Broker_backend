package fixationcustomer

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/infra/repositories/postgres"
	"context"
)

type Repository struct {
	tx *postgres.TxManager
}

func NewRepository(tx *postgres.TxManager) *Repository {
	return &Repository{tx: tx}
}

func (r *Repository) Insert(ctx context.Context, fixationCustomer entity.FixationCustomer) error {
	const op = "postgres.features.Insert"
	ctx, cancel := r.tx.WriteWithTimeout(ctx)
	defer cancel()

	query := `INSERT INTO fixation_customers(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, customer_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
	//верно понял что надо перед вставкой проверить что фиксации нет? как это правильно сделать - сделать метод check в юзкейсах отдельный который будет делать select или сделать в этом методе select и insert транзакцией?
	_, err := r.tx.Querier(ctx).Exec(ctx, query, fixationCustomer.FixationID, fixationCustomer.FixedAt, fixationCustomer.ExpiresAt, fixationCustomer.Status, fixationCustomer.BrokerID, fixationCustomer.FixedBy, fixationCustomer.FixFor, fixationCustomer.CustomerID)
	if err != nil {
		return postgres.MapError(op, err)
	}

	return nil
}

//func (r *Repository) UpdateFixationCustomer(ctx context.Context, brokerID entity.BrokerID, userID entity.UserID, customerID entity.CustomerID, now *time.Time) error { //как сделать чтобы юзкейс видел вызов этого метода? как принято делать?
//	const op = "postgres.features.UpdateFixationCustomer" //есть ли смысл в моем случае писать op? стоит убрать его из мапера и в целом не использовать их?
//	ctx, cancel := r.pg.WriteWithTimeout(ctx)
//	defer cancel()
//
//	query := `UPDATE customers SET broker_id = $1, user_id = $2, updated_at = $3 WHERE customer_id = $4`
//
//	tag, err := r.pg.DB().Exec(ctx, query, brokerID, userID, now, customerID) //не понимаю почему r.pg.DB() а не просто r.pg, r.pg же и есть *postgres.Postgres, почему вот так у меня и правильно ли так делать и почему?
//	if tag.RowsAffected() == 0{
//return domain.ErrNoRows
//}	if err != nil {
//		return postgres.MapError(op, err)
//	}
//
//	return nil
//}
