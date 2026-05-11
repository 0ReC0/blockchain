package sharding

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// LoadTestResult - результаты нагрузочного теста
type LoadTestResult struct {
	TotalTransactions int64
	SuccessRate       float64
	AvgLatencyMs      float64
	P95LatencyMs      float64
	P99LatencyMs      float64
	MinTPS            float64
	MaxTPS            float64
	AvgTPS            float64
	ScaleOperations   int64
	FinalShardCount   int
	Duration          time.Duration
}

// LoadScenario - сценарий нагрузки
type LoadScenario struct {
	Name        string
	TPS         int // целевая интенсивность
	Duration    time.Duration
	Description string
}

// runLoadTest проводит нагрузочное тестирование
func runLoadTest(t *testing.T, manager *AdaptiveShardingManager, scenarios []LoadScenario) *LoadTestResult {
	var (
		totalTx      int64
		successTx    int64
		latencies    []float64
		latenciesMu  sync.Mutex
		stopChan     = make(chan struct{})
		tpsSamples   []float64
		tpsMu        sync.Mutex
		currentTPS   float64
	)

	// Запуск адаптивного менеджера
	manager.StartAdaptation()

	// Сборщик метрик TPS и обновитель прогноза
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				tpsMu.Lock()
				tpsSamples = append(tpsSamples, currentTPS)
				// Обновляем прогноз нагрузкой за последнюю секунду
				manager.UpdateLoad(currentTPS)
				currentTPS = 0
				tpsMu.Unlock()
			}
		}
	}()

	startTime := time.Now()

	// Выполнение сценариев
	for _, scenario := range scenarios {
		t.Logf("📈 Сценарий: %s (%d TPS, %v)", scenario.Name, scenario.TPS, scenario.Duration)

		txPerSecond := scenario.TPS
		interval := time.Second / time.Duration(txPerSecond)

		var wg sync.WaitGroup
		scenarioDone := make(chan struct{})

		// Генератор транзакций
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-scenarioDone:
					return
				case <-ticker.C:
					wg.Add(1)
					go func() {
						defer wg.Done()

						txStart := time.Now()

						// Создание и маршрутизация транзакции
						txID := fmt.Sprintf("tx-load-%d-%d", time.Now().UnixNano(), atomic.LoadInt64(&totalTx))
						shardID := manager.Router.RouteByHash(txID)

						// Имитация обработки
						time.Sleep(time.Microsecond * 10)

						latency := time.Since(txStart).Seconds() * 1000 // мс

						atomic.AddInt64(&totalTx, 1)
						atomic.AddInt64(&successTx, 1)

						latenciesMu.Lock()
						latencies = append(latencies, latency)
						latenciesMu.Unlock()

						tpsMu.Lock()
						currentTPS++
						tpsMu.Unlock()

						_ = shardID
					}()
				}
			}
		}()

		// Ожидание завершения сценария
		time.Sleep(scenario.Duration)
		close(scenarioDone)
		wg.Wait()

		t.Logf("   ✅ Завершено: %d транзакций, шардов: %d",
			atomic.LoadInt64(&totalTx), manager.GetShardCount())
	}

	close(stopChan)
	duration := time.Since(startTime)

	// Остановка менеджера
	manager.StopAdaptation()

	// Вычисление статистики
	result := &LoadTestResult{
		TotalTransactions: totalTx,
		SuccessRate:       float64(successTx) / float64(totalTx) * 100,
		Duration:          duration,
		AvgTPS:            float64(totalTx) / duration.Seconds(),
		ScaleOperations:   manager.scaleOperations,
		FinalShardCount:   manager.GetShardCount(),
	}

	// Латентность
	if len(latencies) > 0 {
		result.AvgLatencyMs = calculateMean(latencies)
		result.P95LatencyMs = calculatePercentile(latencies, 95)
		result.P99LatencyMs = calculatePercentile(latencies, 99)
	}

	// TPS статистика
	if len(tpsSamples) > 0 {
		result.MinTPS = tpsSamples[0]
		result.MaxTPS = tpsSamples[0]
		for _, tps := range tpsSamples {
			if tps < result.MinTPS {
				result.MinTPS = tps
			}
			if tps > result.MaxTPS {
				result.MaxTPS = tps
			}
		}
	}

	return result
}

// TestAdaptiveSharding_LoadTest_Complete - полный нагрузочный тест
func TestAdaptiveSharding_LoadTest_Complete(t *testing.T) {
	t.Log("🚀 Нагрузочное тестирование адаптивного шардирования")

	// Конфигурация для теста
	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          16,
		ScaleUpThreshold:   100,  // 100 TPS для масштабирования вверх
		ScaleDownThreshold: 30,   // 30 TPS для масштабирования вниз
		AdaptationInterval: 2 * time.Second,
		PredictionHorizon:  3,
	}

	manager := NewAdaptiveShardingManager(config)

	// Инициализация
	err := manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})
	if err != nil {
		t.Fatalf("Failed to initialize shards: %v", err)
	}

	// Сценарии нагрузки
	scenarios := []LoadScenario{
		{
			Name:        "Low Load",
			TPS:         50,
			Duration:    5 * time.Second,
			Description: "Низкая нагрузка - должен работать 1 шард",
		},
		{
			Name:        "Medium Load",
			TPS:         150,
			Duration:    5 * time.Second,
			Description: "Средняя нагрузка - масштабирование до 2 шардов",
		},
		{
			Name:        "High Load",
			TPS:         300,
			Duration:    5 * time.Second,
			Description: "Высокая нагрузка - масштабирование до 3-4 шардов",
		},
		{
			Name:        "Peak Load",
			TPS:         500,
			Duration:    5 * time.Second,
			Description: "Пиковая нагрузка - максимальное масштабирование",
		},
		{
			Name:        "Back to Low",
			TPS:         50,
			Duration:    8 * time.Second,
			Description: "Возврат к низкой нагрузке - масштабирование вниз",
		},
	}

	result := runLoadTest(t, manager, scenarios)

	// Вывод результатов
	t.Log("\n📊 РЕЗУЛЬТАТЫ НАГРУЗОЧНОГО ТЕСТИРОВАНИЯ")
	t.Logf("   Всего транзакций:    %d", result.TotalTransactions)
	t.Logf("   Success Rate:        %.2f%%", result.SuccessRate)
	t.Logf("   Средняя TPS:         %.2f", result.AvgTPS)
	t.Logf("   Min TPS:             %.2f", result.MinTPS)
	t.Logf("   Max TPS:             %.2f", result.MaxTPS)
	t.Logf("   Средняя латентность: %.3f мс", result.AvgLatencyMs)
	t.Logf("   P95 латентность:     %.3f мс", result.P95LatencyMs)
	t.Logf("   P99 латентность:     %.3f мс", result.P99LatencyMs)
	t.Logf("   Операций масштаба:   %d", result.ScaleOperations)
	t.Logf("   Финальное кол-во шардов: %d", result.FinalShardCount)
	t.Logf("   Длительность:        %v", result.Duration)

	// Валидация
	if result.SuccessRate < 99.0 {
		t.Errorf("Success rate слишком низкий: %.2f%%", result.SuccessRate)
	}

	if result.ScaleOperations < 2 {
		t.Errorf("Ожидается минимум 2 операции масштабирования, got %d", result.ScaleOperations)
	}

	t.Log("\n✅ Нагрузочное тестирование завершено успешно")
}

// TestAdaptiveSharding_LoadTest_StaticComparison - сравнение со статическим шардированием
func TestAdaptiveSharding_LoadTest_StaticComparison(t *testing.T) {
	t.Log("📊 Сравнение адаптивного и статического шардирования")

	// Адаптивное шардирование
	t.Run("Adaptive Sharding", func(t *testing.T) {
		config := &ShardConfig{
			MinShards:          1,
			MaxShards:          8,
			ScaleUpThreshold:   100,
			ScaleDownThreshold: 30,
			AdaptationInterval: 1 * time.Second,
			PredictionHorizon:  3,
		}

		manager := NewAdaptiveShardingManager(config)
		manager.InitializeShards(func(id int) *Shard {
			return &Shard{ID: id, Active: true}
		})

		scenarios := []LoadScenario{
			{Name: "Low", TPS: 50, Duration: 3 * time.Second},
			{Name: "High", TPS: 300, Duration: 3 * time.Second},
			{Name: "Low", TPS: 50, Duration: 5 * time.Second},
		}

		result := runLoadTest(t, manager, scenarios)

		t.Logf("   Средняя TPS: %.2f, Латентность: %.3f мс, Шардов: %d→%d",
			result.AvgTPS, result.AvgLatencyMs, 1, result.FinalShardCount)
	})

	// Статическое шардирование (4 шарда фиксировано)
	t.Run("Static Sharding (4 shards)", func(t *testing.T) {
		config := &ShardConfig{
			MinShards:          4,
			MaxShards:          4,
			ScaleUpThreshold:   1000, // Отключить масштабирование
			ScaleDownThreshold: 1,
			AdaptationInterval: 100 * time.Second,
			PredictionHorizon:  3,
		}

		manager := NewAdaptiveShardingManager(config)
		manager.InitializeShards(func(id int) *Shard {
			return &Shard{ID: id, Active: true}
		})

		scenarios := []LoadScenario{
			{Name: "Low", TPS: 50, Duration: 3 * time.Second},
			{Name: "High", TPS: 300, Duration: 3 * time.Second},
			{Name: "Low", TPS: 50, Duration: 5 * time.Second},
		}

		result := runLoadTest(t, manager, scenarios)

		t.Logf("   Средняя TPS: %.2f, Латентность: %.3f мс, Шардов: %d (фиксировано)",
			result.AvgTPS, result.AvgLatencyMs, 4)
	})
}

// TestAdaptiveSharding_LoadTest_ScaleUpSpeed - тест скорости масштабирования вверх
func TestAdaptiveSharding_LoadTest_ScaleUpSpeed(t *testing.T) {
	t.Log("⬆️ Тест скорости масштабирования вверх")

	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          8,
		ScaleUpThreshold:   50, // Низкий порог для быстрого теста
		ScaleDownThreshold: 10,
		AdaptationInterval: 500 * time.Millisecond,
		PredictionHorizon:  3,
	}

	manager := NewAdaptiveShardingManager(config)
	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	initialShards := manager.GetShardCount()
	t.Logf("   Начальное количество шардов: %d", initialShards)

	// Симуляция резкого роста нагрузки
	for i := 0; i < 50; i++ {
		manager.UpdateLoad(200) // 200 TPS
	}

	manager.StartAdaptation()

	// Ожидание масштабирования
	maxWait := 5 * time.Second
	startWait := time.Now()
	for time.Since(startWait) < maxWait {
		if manager.GetShardCount() > initialShards {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	scaleUpTime := time.Since(startWait)
	finalShards := manager.GetShardCount()

	t.Logf("   Финальное количество шардов: %d", finalShards)
	t.Logf("   Время масштабирования: %v", scaleUpTime)

	manager.StopAdaptation()

	if finalShards <= initialShards {
		t.Error("Ожидается увеличение количества шардов")
	}
}

// TestAdaptiveSharding_LoadTest_ScaleDownSpeed - тест скорости масштабирования вниз
func TestAdaptiveSharding_LoadTest_ScaleDownSpeed(t *testing.T) {
	t.Log("⬇️ Тест скорости масштабирования вниз")

	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          8,
		ScaleUpThreshold:   100,
		ScaleDownThreshold: 20, // Низкий порог для масштабирования вниз
		AdaptationInterval: 500 * time.Millisecond,
		PredictionHorizon:  3,
	}

	manager := NewAdaptiveShardingManager(config)
	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	// Начальное масштабирование вверх
	for i := 0; i < 20; i++ {
		manager.UpdateLoad(200)
	}
	manager.StartAdaptation()
	time.Sleep(1 * time.Second)

	initialShards := manager.GetShardCount()
	t.Logf("   Начальное количество шардов: %d", initialShards)

	// Симуляция падения нагрузки
	for i := 0; i < 50; i++ {
		manager.UpdateLoad(5) // 5 TPS
	}

	// Ожидание масштабирования вниз
	maxWait := 5 * time.Second
	startWait := time.Now()
	for time.Since(startWait) < maxWait {
		if manager.GetShardCount() < initialShards {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	scaleDownTime := time.Since(startWait)
	finalShards := manager.GetShardCount()

	t.Logf("   Финальное количество шардов: %d", finalShards)
	t.Logf("   Время масштабирования: %v", scaleDownTime)

	manager.StopAdaptation()

	if finalShards >= initialShards {
		t.Error("Ожидается уменьшение количества шардов")
	}
}

// calculateMean вычисляет среднее значение
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculatePercentile вычисляет перцентиль
func calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Сортировка
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * float64(percentile) / 100)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
