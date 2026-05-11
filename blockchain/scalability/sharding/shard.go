package sharding

import (
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
	"fmt"
	"sync"
	"time"
)

// Шард
type Shard struct {
	ID         int
	Validators []string
	Chain      *blockchain.Blockchain
	TxPool     *txpool.TransactionPool
	Active     bool // флаг активности шарда (для динамического добавления/удаления)
}

// ShardConfig - конфигурация для адаптивного шардирования
type ShardConfig struct {
	MinShards       int           // минимальное количество шардов
	MaxShards       int           // максимальное количество шардов
	ScaleUpThreshold   float64     // порог нагрузки для добавления шарда (TPS)
	ScaleDownThreshold float64     // порог нагрузки для удаления шарда (TPS)
	AdaptationInterval time.Duration // интервал адаптации конфигурации
	PredictionHorizon  int           // горизонт прогнозирования (шагов)
}

// DefaultShardConfig возвращает конфигурацию по умолчанию
func DefaultShardConfig() *ShardConfig {
	return &ShardConfig{
		MinShards:          1,
		MaxShards:          16,
		ScaleUpThreshold:   1000,  // 1000 TPS для масштабирования вверх
		ScaleDownThreshold: 300,   // 300 TPS для масштабирования вниз
		AdaptationInterval: 30 * time.Second,
		PredictionHorizon:  5,
	}
}

// AdaptiveShardingManager - менеджер адаптивного шардирования
type AdaptiveShardingManager struct {
	mu sync.RWMutex

	// Базовый менеджер
	Shards map[int]*Shard
	Router *ShardRouter

	// Прогнозирование
	Predictor *ARIMAPredictor

	// Конфигурация
	Config *ShardConfig

	// Состояние
	currentShardCount int
	isAdaptive        bool
	stopChan          chan struct{}
	wg                sync.WaitGroup

	// Метрики
	lastAdaptationTime time.Time
	scaleOperations    int64
}

// NewAdaptiveShardingManager создаёт новый менеджер адаптивного шардирования
func NewAdaptiveShardingManager(config *ShardConfig) *AdaptiveShardingManager {
	if config == nil {
		config = DefaultShardConfig()
	}

	manager := &AdaptiveShardingManager{
		Shards:              make(map[int]*Shard),
		Router:              &ShardRouter{ShardCount: config.MinShards},
		Predictor:           NewARIMAPredictor(20), // окно из 20 наблюдений
		Config:              config,
		currentShardCount:   config.MinShards,
		isAdaptive:          true,
		stopChan:            make(chan struct{}),
		lastAdaptationTime:  time.Now(),
	}

	return manager
}

// InitializeShards инициализирует начальное количество шардов
func (m *AdaptiveShardingManager) InitializeShards(createShardFunc func(id int) *Shard) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < m.Config.MinShards; i++ {
		shard := createShardFunc(i)
		shard.Active = true
		m.Shards[i] = shard
	}

	m.Router.ShardCount = m.Config.MinShards
	m.currentShardCount = m.Config.MinShards

	return nil
}

// StartAdaptation запускает цикл адаптивной настройки количества шардов
func (m *AdaptiveShardingManager) StartAdaptation() {
	m.wg.Add(1)
	go m.adaptationLoop()
	fmt.Printf("🔄 Adaptive sharding started: min=%d, max=%d, scaleUp=%.0f TPS, scaleDown=%.0f TPS\n",
		m.Config.MinShards, m.Config.MaxShards,
		m.Config.ScaleUpThreshold, m.Config.ScaleDownThreshold)
}

// StopAdaptation останавливает адаптивный цикл
func (m *AdaptiveShardingManager) StopAdaptation() {
	close(m.stopChan)
	m.wg.Wait()
	fmt.Println("⏹️ Adaptive sharding stopped")
}

// adaptationLoop - основной цикл адаптации
func (m *AdaptiveShardingManager) adaptationLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.Config.AdaptationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.adaptShardCount()
		case <-m.stopChan:
			return
		}
	}
}

// adaptShardCount адаптирует количество шардов на основе прогноза нагрузки
func (m *AdaptiveShardingManager) adaptShardCount() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Получаем прогноз нагрузки
	predictedTPS := m.Predictor.GetPredictedTPS(m.Config.PredictionHorizon)
	loadLevel := m.Predictor.GetLoadLevel()

	fmt.Printf("📊 Load prediction: %.2f TPS (%s) | Current shards: %d\n",
		predictedTPS, loadLevel, m.currentShardCount)

	targetShards := m.calculateTargetShards(predictedTPS)

	if targetShards != m.currentShardCount {
		m.scaleShards(targetShards)
	}

	m.lastAdaptationTime = time.Now()
}

// calculateTargetShards вычисляет целевое количество шардов на основе прогноза
func (m *AdaptiveShardingManager) calculateTargetShards(predictedTPS float64) int {
	// Базовый расчёт: 1 шард на ScaleUpThreshold TPS
	target := int(predictedTPS/m.Config.ScaleUpThreshold) + 1

	// Ограничиваем минимальным и максимальным значением
	if target < m.Config.MinShards {
		target = m.Config.MinShards
	}
	if target > m.Config.MaxShards {
		target = m.Config.MaxShards
	}

	return target
}

// scaleShards масштабирует количество шардов до целевого значения
func (m *AdaptiveShardingManager) scaleShards(targetCount int) {
	current := m.currentShardCount

	if targetCount > current {
		// Масштабирование вверх
		fmt.Printf("⬆️ Scaling UP: %d -> %d shards (predicted load: %.2f TPS)\n",
			current, targetCount, m.Predictor.GetPredictedTPS(m.Config.PredictionHorizon))
		m.addShards(targetCount - current)
	} else if targetCount < current {
		// Масштабирование вниз
		fmt.Printf("⬇️ Scaling DOWN: %d -> %d shards (predicted load: %.2f TPS)\n",
			current, targetCount, m.Predictor.GetPredictedTPS(m.Config.PredictionHorizon))
		m.removeShards(current - targetCount)
	}

	m.currentShardCount = targetCount
	m.Router.ShardCount = targetCount
	m.scaleOperations++
}

// addShards добавляет новые шарды
func (m *AdaptiveShardingManager) addShards(count int) {
	// Находим следующий доступный ID
	nextID := 0
	for id := range m.Shards {
		if id >= nextID {
			nextID = id + 1
		}
	}

	for i := 0; i < count; i++ {
		// Создаём новый шард с использованием функции извне
		// В реальной реализации здесь будет вызов callback-функции
		shard := &Shard{
			ID:         nextID + i,
			Active:     true,
			Validators: []string{},
		}
		m.Shards[shard.ID] = shard
		fmt.Printf("  ➕ Added shard %d\n", shard.ID)
	}
}

// removeShards удаляет лишние шарды
func (m *AdaptiveShardingManager) removeShards(count int) {
	// Находим неактивные или наименее загруженные шарды для удаления
	removed := 0
	for id, shard := range m.Shards {
		if removed >= count {
			break
		}

		// Не удаляем последние активные шарды
		if len(m.Shards)-removed <= m.Config.MinShards {
			break
		}

		// Помечаем шард как неактивный
		shard.Active = false
		delete(m.Shards, id)
		fmt.Printf("  ➖ Removed shard %d\n", id)
		removed++
	}
}

// RecordTransaction регистрирует транзакцию для обновления прогноза нагрузки
func (m *AdaptiveShardingManager) RecordTransaction(tx *txpool.Transaction) {
	// Обновляем прогноз на основе каждой транзакции
	// В реальной системе можно агрегировать транзакции за интервал времени
	m.Predictor.Update(1.0) // Каждая транзакция = 1 единица нагрузки
}

// UpdateLoad обновляет метрику нагрузки (TPS за интервал)
func (m *AdaptiveShardingManager) UpdateLoad(tps float64) {
	m.Predictor.Update(tps)
}

// GetStats возвращает статистику адаптивного шардирования
func (m *AdaptiveShardingManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mean, variance, trend := m.Predictor.GetStats()

	return map[string]interface{}{
		"current_shards":     m.currentShardCount,
		"min_shards":         m.Config.MinShards,
		"max_shards":         m.Config.MaxShards,
		"predicted_tps":      m.Predictor.GetPredictedTPS(m.Config.PredictionHorizon),
		"load_level":         m.Predictor.GetLoadLevel().String(),
		"mean_load":          mean,
		"variance":           variance,
		"trend":              trend,
		"last_adaptation":    m.lastAdaptationTime,
		"scale_operations":   m.scaleOperations,
		"is_adaptive":        m.isAdaptive,
	}
}

// GetActiveShards возвращает список активных шардов
func (m *AdaptiveShardingManager) GetActiveShards() []*Shard {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*Shard, 0, m.currentShardCount)
	for _, shard := range m.Shards {
		if shard.Active {
			active = append(active, shard)
		}
	}
	return active
}

// GetShardCount возвращает текущее количество шардов
func (m *AdaptiveShardingManager) GetShardCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentShardCount
}

// EnableAdaptivity включает/выключает адаптивный режим
func (m *AdaptiveShardingManager) EnableAdaptivity(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isAdaptive = enabled
	if enabled {
		fmt.Println("✅ Adaptive sharding ENABLED")
	} else {
		fmt.Println("⏸️ Adaptive sharding DISABLED")
	}
}