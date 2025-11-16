package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	baseURL    = "http://localhost:" + envOrDefault("APP_PORT", "8080")
	adminToken = envOrDefault("ADMIN_TOKEN", "secret_admin_token")
)

// маленький helper для JSON-запросов с проверкой статуса и парсингом ответа.
func doJSON(t *testing.T, method, path string, body any, expectedStatus int, out any) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("не удалось закодировать body: %v", err)
		}
	}

	req, err := http.NewRequest(method, baseURL+path, &buf)
	if err != nil {
		t.Fatalf("не удалось создать запрос %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("ошибка при выполнении запроса %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		var respBody bytes.Buffer
		_, _ = respBody.ReadFrom(resp.Body)
		t.Fatalf(
			"ожидали статус %d, получили %d, тело ответа: %s",
			expectedStatus,
			resp.StatusCode,
			respBody.String(),
		)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("не удалось распарсить ответ: %v", err)
		}
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
