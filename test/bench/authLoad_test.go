package bench

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func BenchmarkLoadTestAuth(b *testing.B) {
	// Настройки нагрузки
	rate := vegeta.Rate{Freq: 1000, Per: time.Second} // 1000 RPS
	duration := 1 * time.Second                       // Тестируем 10 секунд

	var targets []vegeta.Target
	for i := 0; i < 1000; i++ { // 100,000 уникальных пользователей
		username := fmt.Sprintf("user%d", i)
		body := fmt.Sprintf(`{"username": "%s", "password": "password"}`, username)
		targets = append(targets, vegeta.Target{
			Method: "POST",
			URL:    "http://localhost:8080/api/auth",
			Header: http.Header{"Content-Type": []string{"application/json"}},
			Body:   []byte(body),
		})
	}

	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for res := range attacker.Attack(targeter, rate, duration, "auth-test") {
			metrics.Add(res)
		}
	}

	metrics.Close()

	// Логирование статистики
	b.Logf("===== РЕЗУЛЬТАТЫ БЕНЧМАРКА =====")
	b.Logf("Запросов: %d", metrics.Requests)
	b.Logf("Среднее время ответа: %.2f ms", metrics.Latencies.Mean.Seconds()*1000)
	b.Logf("99-й перцентиль времени ответа: %.2f ms", metrics.Latencies.P99.Seconds()*1000)
	b.Logf("Успешные запросы: %.2f%%", metrics.Success*100)
	b.Logf("Средняя скорость: %.2f RPS", metrics.Rate)

	// Проверка на соответствие требованиям
	if metrics.Success < 0.9999 {
		b.Error("❌ Ошибка: успешность ниже 99.99%")
	} else {
		b.Log("✅ Успешность выше 99.99%")
	}

	if metrics.Latencies.P99.Seconds()*1000 > 50 {
		b.Error("❌ Ошибка: время ответа P99 больше 50 мс")
	} else {
		b.Log("✅ Время ответа соответствует ≤ 50 мс")
	}
}
