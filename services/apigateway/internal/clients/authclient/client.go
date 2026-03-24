package authclient

import (
	"Donate_backend/services/apigateway/internal/configs"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"
	"context"

	"google.golang.org/grpc"
)

type Client struct {
	auth   authv1.AuthServiceClient
	config *configs.Config
}

func NewAuthServiceClient() authv1.AuthServiceClient { //TODO: так всё теперь? не понял если честно даже на уровне идеи зачем мы создаем экземпляр интерфейса.
	// В моем понимании у интерфейсов достаточно просто сигнатуры, они нужны в 99% просто чтобы указать сигнатуру методов в них и подменить интерфейс
	//какой-то структурой. А тут мы сгенерировали сигнатуру и пошли создавать экземпляр интерфейса. А дальше вместо этого интерфейса мы
	//подставляем структуру в authservice authv1.UnimplenetedAuthServiceServer(это я написал чтобы себя проверить верно ли суть понял с
	//прошлого вопроса по этой части). То есть мы подменяем этот интерфейс, но зачем создаем экземпляр интерфейса неясно - если честно тут до
	//сих пор не понял роль интерфейса и структуры authv1.Unimplented
	conn, err := grpc.Dial() //TODO: тут тоже не понял что делает и зачем
	return authv1.NewAuthServiceClient(conn)
}

func NewClient(auth authv1.AuthServiceClient, cfg *configs.Config) *Client {
	return &Client{auth: auth, config: cfg}
}

func (c *Client) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) { //TODO: как чаще на high load проектах в рф делают - хендлер преобразовывает dto в grpcReq или transport в этой функции? наверно логично в хендлере потому что он и нужен для пасринга и преобразований, а транспорт отвечает просто за отправку, так?
	ctx, cancel := context.WithTimeout(ctx, c.config.ContextTimeout)
	defer cancel()

	return c.auth.Register(ctx, req)
}
