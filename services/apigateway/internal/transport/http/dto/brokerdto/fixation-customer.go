package brokerdto

type FixationCustomerRequest struct { //если на шаг назад - dto это место куда парсим body или место куда парсим и body и квери параметры?
	BrokerID   string `json:"broker_id" validate:"required,uuid6"`
	CustomerID string `json:"customer_id" validate:"required,uuid6"` //мешают ли для мапингу приватные типы, если да - почему?
	FixFor     string `json:"fix_for" validate:"required,fix_for"`
	//какое вообще правило - когда кладу в бади а когда в квери параметры? вот например есть customer_id - мне его куда класть?
}

type FixationCustomerResponse struct {
	FixationID string `json:"fixation_id" validate:"required,uuid6"` //тут будет срабатывать валидация? смущает что в хендлерах если верно понимаю, срабатывает только один раз в начале валидатор, покажи прав ли я, миддлвар или другое что-то у меня валидацию делает
	Status     string `json:"status" validate:"required,status"`
	FixedAt    string `json:"fixed_at" validate:"required,fixed_at"`
	ExpiresAt  string `json:"expires_at" validate:"required,expires_at"`
}
