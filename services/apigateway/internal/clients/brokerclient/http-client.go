package brokerclient

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Client interface { //я в очередной раз не понимаю как это применять и нормально подменять. напиши мне по шагам. где интерфейс, где структура которая его заменяет и зачем она и интерфейс вообще тут. в очередной раз не понимаю что писать, в очередной раз пытаюсь пойти на уровень ниже и вижу что не понимаю что пишу тут, что это даст, зачем. я помню правило что интерфейс пишем по месту применения, что нужен он чтобы можно было мокать автотест. тут у меня слой клиент гейтвея, клиент просто шина которая ничего не делает в моем понимании(да устанавливает соединение еще, но логики или изменения данных тут нет), нужен ли интерфейс? и автотесты нужны ли именно тут?
	//то есть вот ситуация - клиенту нужно вызвать слой ниже, транспорт брокерсервиса. Как называю интерфейс? метод который в нем - метод слоя который ниже или метод из текущего слоя где интерфейс? видимо в этом главное не понимание и плюс надо делать тест тут или нет(то есть нужен ли вообще тут интерфейс) и кстати, когда мы можем отправить и grpc и по http, у нас методы должны лежать в общем интерфейсе или лучше разделить на два файла как у меня сейчас?
	UpdateFixation(ctx context.Context, req *brokerdto.FixationRequest, brokerID, userID uuid.UUID) (*brokerdto.FixationResponse, error)
}

type HTTPClient struct {
	httpClient *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{},
	}
}

func (c *HTTPClient) UpdateFixation(ctx context.Context, req *brokerdto.FixationRequest, brokerID, userID uuid.UUID) (*brokerdto.FixationResponse, error) {
	return c.httpClient.UpdateFixation(ctx, req, brokerID, userID)
}
