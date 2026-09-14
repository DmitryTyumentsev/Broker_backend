package amocrm

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"bytes"
	"context"
	"net/http"
)

const (
	eventFixationCreated = "fixation.created"
)

func (s *Sender) Send(ctx context.Context, events []entity.Outbox) error {
	ctx, cancel := context.WithTimeout(ctx, s.Config.Integrations.AmoCRM.Timeout)
	defer cancel()

	for _, ev := range events {
		switch ev.EventType {
		case eventFixationCreated:
			err := s.sendFixationCreated(ctx, ev)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Sender) sendFixationCreated(ctx context.Context, ev entity.Outbox) error {
	url := s.Config.Integrations.AmoCRM.EndpointSendLeads
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(ev)))
	if err != nil {
		return err
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer _=resp.Body.Close() //зачем я контекст сюда в метод передаю?

	if resp.StatusCode != http.StatusOK || resp.StatusCode != http.StatusCreated {
		return err
	}
	return nil
}
