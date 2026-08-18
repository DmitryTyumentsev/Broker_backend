// token — логинится в partnerapi и печатает access-токен.
//
//	go run ./tools/token -email broker1@perviy-metr.test
//	TOKEN=$(make token EMAIL=broker1@perviy-metr.test)
//
// Зачем отдельный бинарь, а не curl с jq: jq есть не у всех, а токен
// нужен в каждой ручной проверке и в racefix. Одна команда вместо
// «скопируй поле access из ответа».
//
// Печатает ТОЛЬКО токен в stdout — чтобы результат можно было
// подставить в переменную. Всё остальное уходит в stderr.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type loginResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "token: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		baseURL  = flag.String("url", "http://localhost:8080", "адрес partnerapi")
		email    = flag.String("email", "", "почта сотрудника (обязательна)")
		password = flag.String("password", "password", "пароль; у сидера он у всех одинаковый")
		deviceID = flag.String("device", "cli", "идентификатор устройства")
		timeout  = flag.Duration("timeout", 15*time.Second, "потолок на запрос")
		verbose  = flag.Bool("v", false, "печатать в stderr ещё и refresh-токен")
	)
	flag.Parse()

	if strings.TrimSpace(*email) == "" {
		flag.Usage()

		return errors.New("нужен -email; список сотрудников печатает `make seed`")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	payload, err := json.Marshal(loginRequest{
		Email:    *email,
		Password: *password,
		DeviceID: *deviceID,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(*baseURL, "/") + "/api/v1/auth/login"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("запрос к %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("чтение ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Тело отдаём как есть: на этом этапе полезнее увидеть настоящий
		// ответ сервиса, чем нашу интерпретацию.
		return fmt.Errorf("partnerapi ответил %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded loginResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("разбор ответа: %w", err)
	}

	if decoded.Access == "" {
		return fmt.Errorf("в ответе нет access-токена: %s", strings.TrimSpace(string(body)))
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "refresh: %s\n", decoded.Refresh)
	}

	fmt.Println(decoded.Access)

	return nil
}
