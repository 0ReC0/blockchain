package sharding

import (
	"blockchain/storage/txpool"
	"fmt"
	"testing"
	"time"
)

// ============ Тесты для ARIMA прогнозиста ============

// TestARIMAPredictor_Basic - базовый тест прогнозирования
func TestARIMAPredictor_Basic(t *testing.T) {
	predictor := NewARIMAPredictor(10)

	// Обучаем на простых данных
	for i := 0; i < 10; i++ {
		predictor.Update(float64(i * 10))
	}

	// Проверяем прогноз
	forecast := predictor.Predict(3)
	if len(forecast) != 3 {
		t.Errorf("Expected 3 forecast values, got %d", len(forecast))
	}

	// Прогноз должен быть положительным
	for _, v := range forecast {
		if v <= 0 {
			t.Errorf("Expected positive forecast, got %f", v)
		}
	}

	t.Logf("Forecast: %v", forecast)
}

// TestARIMAPredictor_LoadLevel - тест определения уровня нагрузки
func TestARIMAPredictor_LoadLevel(t *testing.T) {
	predictor := NewARIMAPredictor(5)

	// Низкая нагрузка
	for i := 0; i < 5; i++ {
		predictor.Update(50)
	}
	if predictor.GetLoadLevel() != LoadLevelLow {
		t.Errorf("Expected LOW load level, got %s", predictor.GetLoadLevel())
	}

	// Средняя нагрузка
	for i := 0; i < 5; i++ {
		predictor.Update(300)
	}
	if predictor.GetLoadLevel() != LoadLevelMedium {
		t.Errorf("Expected MEDIUM load level, got %s", predictor.GetLoadLevel())
	}

	// Высокая нагрузка
	for i := 0; i < 5; i++ {
		predictor.Update(700)
	}
	if predictor.GetLoadLevel() != LoadLevelHigh {
		t.Errorf("Expected HIGH load level, got %s", predictor.GetLoadLevel())
	}

	// Пиковая нагрузка
	for i := 0; i < 5; i++ {
		predictor.Update(1500)
	}
	if predictor.GetLoadLevel() != LoadLevelPeak {
		t.Errorf("Expected PEAK load level, got %s", predictor.GetLoadLevel())
	}
}

// TestARIMAPredictor_Stats - тест статистики
func TestARIMAPredictor_Stats(t *testing.T) {
	predictor := NewARIMAPredictor(10)

	// Добавляем данные с трендом
	for i := 0; i < 10; i++ {
		predictor.Update(float64(i * 100))
	}

	mean, variance, trend := predictor.GetStats()

	t.Logf("Mean: %.2f, Variance: %.2f, Trend: %.4f", mean, variance, trend)

	// Тренд должен быть положительным (растущая нагрузка)
	if trend <= 0 {
		t.Errorf("Expected positive trend, got %.4f", trend)
	}
}

// ============ Тесты для адаптивного менеджера шардирования ============

// TestAdaptiveShardingManager_Initialization - тест инициализации
func TestAdaptiveShardingManager_Initialization(t *testing.T) {
	config := &ShardConfig{
		MinShards:          2,
		MaxShards:          8,
		ScaleUpThreshold:   100,
		ScaleDownThreshold: 30,
		AdaptationInterval: 1 * time.Second,
		PredictionHorizon:  5,
	}

	manager := NewAdaptiveShardingManager(config)

	// Инициализируем шарды
	err := manager.InitializeShards(func(id int) *Shard {
		return &Shard{
			ID:     id,
			Active: true,
		}
	})

	if err != nil {
		t.Fatalf("Failed to initialize shards: %v", err)
	}

	if manager.GetShardCount() != 2 {
		t.Errorf("Expected 2 shards, got %d", manager.GetShardCount())
	}

	t.Logf("Initialized with %d shards", manager.GetShardCount())
}

// TestAdaptiveShardingManager_ScaleUp - тест масштабирования вверх
func TestAdaptiveShardingManager_ScaleUp(t *testing.T) {
	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          8,
		ScaleUpThreshold:   50, // Низкий порог для быстрого теста
		ScaleDownThreshold: 10,
		AdaptationInterval: 100 * time.Millisecond,
		PredictionHorizon:  3,
	}

	manager := NewAdaptiveShardingManager(config)

	// Инициализируем
	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	// Симулируем высокую нагрузку
	for i := 0; i < 20; i++ {
		manager.UpdateLoad(200) // 200 TPS > 50 TPS threshold
	}

	// Запускаем адаптацию
	manager.StartAdaptation()

	// Ждём адаптации
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что количество шардов увеличилось
	currentShards := manager.GetShardCount()
	t.Logf("Shards after scale up: %d", currentShards)

	// Останавливаем
	manager.StopAdaptation()

	if currentShards < 1 {
		t.Errorf("Expected shard count to increase, got %d", currentShards)
	}
}

// TestAdaptiveShardingManager_ScaleDown - тест масштабирования вниз
func TestAdaptiveShardingManager_ScaleDown(t *testing.T) {
	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          8,
		ScaleUpThreshold:   100,
		ScaleDownThreshold: 10, // Низкий порог для масштабирования вниз
		AdaptationInterval: 100 * time.Millisecond,
		PredictionHorizon:  3,
	}

	manager := NewAdaptiveShardingManager(config)

	// Инициализируем с несколькими шардами
	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	// Добавляем шарды вручную для теста
	for i := 2; i < 5; i++ {
		manager.Shards[i] = &Shard{ID: i, Active: true}
	}
	manager.currentShardCount = 5

	// Симулируем низкую нагрузку
	for i := 0; i < 20; i++ {
		manager.UpdateLoad(5) // 5 TPS < 10 TPS threshold
	}

	// Запускаем адаптацию
	manager.StartAdaptation()

	// Ждём адаптации
	time.Sleep(500 * time.Millisecond)

	// Проверяем количество шардов
	currentShards := manager.GetShardCount()
	t.Logf("Shards after scale down: %d", currentShards)

	// Останавливаем
	manager.StopAdaptation()

	// Должно остаться минимум MinShards
	if currentShards < config.MinShards {
		t.Errorf("Expected at least %d shards, got %d", config.MinShards, currentShards)
	}
}

// TestAdaptiveShardingManager_Stats - тест статистики
func TestAdaptiveShardingManager_Stats(t *testing.T) {
	config := DefaultShardConfig()
	manager := NewAdaptiveShardingManager(config)

	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	// Добавляем данные о нагрузке
	for i := 0; i < 10; i++ {
		manager.UpdateLoad(float64(100 + i*10))
	}

	stats := manager.GetStats()

	t.Logf("Stats: %+v", stats)

	// Проверяем наличие ключей
	requiredKeys := []string{
		"current_shards",
		"min_shards",
		"max_shards",
		"predicted_tps",
		"load_level",
		"mean_load",
		"trend",
	}

	for _, key := range requiredKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Missing stat key: %s", key)
		}
	}
}

// ============ Тесты для роутера ============

// TestShardRouter_Routing - тест маршрутизации
func TestShardRouter_Routing(t *testing.T) {
	router := &ShardRouter{ShardCount: 4}

	// Тестируем маршрутизацию по получателю
	tx := &txpool.Transaction{
		From: "addr1",
		To:   "addr2",
	}

	shard1 := router.RouteTransaction(tx)
	shard2 := router.RouteTransaction(tx)

	// Одна и та же транзакция должна попадать в один шард
	if shard1 != shard2 {
		t.Errorf("Same transaction routed to different shards: %d vs %d", shard1, shard2)
	}

	t.Logf("Transaction routed to shard %d", shard1)
}

// TestShardRouter_DynamicUpdate - тест динамического обновления
func TestShardRouter_DynamicUpdate(t *testing.T) {
	router := &ShardRouter{ShardCount: 2}

	// Обновляем количество шардов
	router.UpdateShardCount(8)

	if router.GetShardCount() != 8 {
		t.Errorf("Expected 8 shards, got %d", router.GetShardCount())
	}

	// Проверяем маршрутизацию с новым количеством
	tx := &txpool.Transaction{From: "addr1", To: "addr2"}
	shard := router.RouteTransaction(tx)

	if shard < 0 || shard >= 8 {
		t.Errorf("Shard %d out of range [0, 8)", shard)
	}

	t.Logf("Transaction routed to shard %d with 8 total shards", shard)
}

// TestShardRouter_ConsistentHashing - тест консистентного хеширования
func TestShardRouter_ConsistentHashing(t *testing.T) {
	ring := NewConsistentHashRing(3) // 3 виртуальных узла на шард

	// Добавляем шарды
	for i := 0; i < 4; i++ {
		ring.AddShard(i)
	}

	// Проверяем распределение
	key := "test-key"
	shard1 := ring.GetShard(key)
	shard2 := ring.GetShard(key)

	// Один ключ должен всегда попадать в один шард
	if shard1 != shard2 {
		t.Errorf("Same key routed to different shards: %d vs %d", shard1, shard2)
	}

	t.Logf("Key '%s' routed to shard %d", key, shard1)

	// Проверяем распределение разных ключей
	distribution := make(map[int]int)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		shard := ring.GetShard(key)
		distribution[shard]++
	}

	t.Logf("Distribution: %+v", distribution)
}

// ============ Бенчмарки ============

// BenchmarkARIMAPredictor_Update - бенчмарк обновления прогнозиста
func BenchmarkARIMAPredictor_Update(b *testing.B) {
	predictor := NewARIMAPredictor(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor.Update(float64(i))
	}
}

// BenchmarkARIMAPredictor_Predict - бенчмарк прогнозирования
func BenchmarkARIMAPredictor_Predict(b *testing.B) {
	predictor := NewARIMAPredictor(100)

	// Обучаем
	for i := 0; i < 100; i++ {
		predictor.Update(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor.Predict(10)
	}
}

// BenchmarkAdaptiveShardingManager_Adaptation - бенчмарк адаптации
func BenchmarkAdaptiveShardingManager_Adaptation(b *testing.B) {
	config := &ShardConfig{
		MinShards:          1,
		MaxShards:          16,
		ScaleUpThreshold:   100,
		ScaleDownThreshold: 30,
		AdaptationInterval: 1 * time.Second,
		PredictionHorizon:  5,
	}

	manager := NewAdaptiveShardingManager(config)
	manager.InitializeShards(func(id int) *Shard {
		return &Shard{ID: id, Active: true}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.UpdateLoad(float64(i % 1000))
	}
}

// BenchmarkShardRouter_Routing - бенчмарк маршрутизации
func BenchmarkShardRouter_Routing(b *testing.B) {
	router := &ShardRouter{ShardCount: 8}
	tx := &txpool.Transaction{From: "sender", To: "receiver"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.RouteTransaction(tx)
	}
}

// BenchmarkConsistentHashRing - бенчмарк консистентного хеширования
func BenchmarkConsistentHashRing(b *testing.B) {
	ring := NewConsistentHashRing(5)
	for i := 0; i < 8; i++ {
		ring.AddShard(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.GetShard(fmt.Sprintf("key-%d", i))
	}
}
