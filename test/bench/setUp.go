package bench

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func GetAuthToken(b *testing.B, username, password string) string {
	body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password)
	req, err := http.NewRequest("POST", "http://localhost:8080/api/auth", nil)
	if err != nil {
		b.Fatalf("Не удалось создать запрос: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(strings.NewReader(body))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		b.Fatalf("Не удалось выполнить запрос: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.Fatalf("Ошибка при получении токена, код %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		b.Fatalf("Не удалось распарсить ответ: %v", err)
	}

	token, ok := result["token"].(string)
	if !ok {
		b.Fatalf("Не удалось извлечь токен из ответа")
	}

	return token
}
