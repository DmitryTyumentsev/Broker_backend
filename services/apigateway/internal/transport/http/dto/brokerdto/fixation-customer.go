package brokerdto

import (
	"time"

	"github.com/google/uuid"
)

type FixationRequest struct { //верно ли на разные методы(new и update например) использовать одни и те же dto добавив поля для совместного использования? например из-за update, добавил в FixationRequest поле FixationIDNew которое не нужно ручке new-fixation
	Phone         string    `json:"phone" validate:"required"`
	FixFor        uuid.UUID `json:"fix_for" validate:"required,uuid"`
	ProjectID     uuid.UUID `json:"project_id" validate:"required,uuid"`
	FixationIDOld uuid.UUID `json:"fixation_id_old" validate:"uuid"`
	//какое вообще правило - когда кладу в бади а когда в квери параметры? и dto = body? квери параметры добавляют в dto или нет? ответишь и тут на оба вопроса
}

type FixationResponse struct {
	FixationIDNew uuid.UUID `json:"fixation_id_new" validate:"required,uuid"` //тут будет срабатывать валидация? смущает что в хендлерах если верно понимаю, срабатывает только один раз в начале валидатор, покажи прав ли я, миддлвар или другое что-то у меня валидацию делает
	//Status        string    `json:"status" validate:"required"`//нужно ли нам для ответа в new и update ручках поле status? убрал его потому что мне кажется оно не нужно в ответе, объясни как принято делать, как правильно
	FixedAt   time.Time `json:"fixed_at" validate:"required"`
	ExpiresAt time.Time `json:"expires_at"`
}
