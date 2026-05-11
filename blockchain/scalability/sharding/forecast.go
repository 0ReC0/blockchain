package sharding

import (
	"math"
	"sync"
)

// LoadPredictor - интерфейс для прогнозирования нагрузки
type LoadPredictor interface {
	Predict(steps int) []float64
	Update(value float64)
}

// ARIMAPredictor - реализация ARIMA(1,1,1) для прогнозирования интенсивности транзакций
// Упрощённая модель для реального времени
type ARIMAPredictor struct {
	mu sync.RWMutex

	// Параметры модели ARIMA(1,1,1)
	phi   float64 // AR коэффициент
	theta float64 // MA коэффициент

	// Внутренние состояния
	lastValue    float64 // последнее наблюдение
	lastDiff     float64 // последняя разность
	lastError    float64 // последняя ошибка прогноза
	mean         float64 // среднее значение
	variance     float64 // дисперсия
	observations int     // количество наблюдений

	// Буфер для сглаживания
	windowSize int
	window     []float64
	windowSum  float64
}

// NewARIMAPredictor создаёт новый прогнозист с параметрами по умолчанию
func NewARIMAPredictor(windowSize int) *ARIMAPredictor {
	return &ARIMAPredictor{
		phi:          0.5,  // коэффициент авторегрессии
		theta:        0.3,  // коэффициент скользящего среднего
		windowSize:   windowSize,
		window:       make([]float64, 0, windowSize),
		observations: 0,
	}
}

// Update обновляет модель новым наблюдением
func (p *ARIMAPredictor) Update(value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Обновление скользящего окна
	if len(p.window) >= p.windowSize {
		// Удаляем старое значение
		p.windowSum -= p.window[0]
		p.window = p.window[1:]
	}
	p.window = append(p.window, value)
	p.windowSum += value

	// Обновление статистик
	p.observations++
	p.mean = p.windowSum / float64(len(p.window))

	// Вычисление разности (дифференцирование для стационарности)
	diff := value - p.lastValue

	// Обновление дисперсии
	if p.observations > 1 {
		delta := value - p.mean
		p.variance = p.variance + (delta*(value-p.mean) - p.variance) / float64(p.observations)
	}

	// Вычисление ошибки прогноза
	predicted := p.predictNext()
	p.lastError = value - predicted

	// Адаптивная настройка параметров (онлайн обучение)
	p.adaptParameters(diff)

	p.lastValue = value
	p.lastDiff = diff
}

// predictNext вычисляет следующее значение на основе текущей модели
func (p *ARIMAPredictor) predictNext() float64 {
	if p.observations < 3 {
		return p.mean
	}

	// ARIMA(1,1,1): y_t = mean + phi * (y_{t-1} - mean) + theta * error_{t-1}
	predicted := p.mean + p.phi*p.lastDiff + p.theta*p.lastError
	return predicted
}

// Predict прогнозирует значения на заданное количество шагов вперёд
func (p *ARIMAPredictor) Predict(steps int) []float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.observations < 3 {
		// Недостаточно данных для прогноза
		result := make([]float64, steps)
		for i := range result {
			result[i] = p.mean
		}
		return result
	}

	forecast := make([]float64, steps)

	// Итеративный прогноз
	lastDiff := p.lastDiff
	lastError := p.lastError

	for i := 0; i < steps; i++ {
		// Прогноз следующего значения
		forecast[i] = p.mean + p.phi*lastDiff + p.theta*lastError

		// Для следующего шага используем прогнозированное значение
		lastDiff = forecast[i] - p.mean
		lastError = 0 // Будущие ошибки неизвестны, предполагаем 0
	}

	return forecast
}

// adaptParameters адаптивно настраивает параметры модели на основе новых данных
func (p *ARIMAPredictor) adaptParameters(diff float64) {
	// Простая адаптация на основе градиентного спуска
	learningRate := 0.01

	// Адаптация phi (AR коэффициент)
	if p.lastDiff != 0 {
		gradientPhi := -p.lastDiff * p.lastError
		p.phi -= learningRate * gradientPhi
		// Ограничение phi в разумных пределах
		p.phi = math.Max(-0.9, math.Min(0.9, p.phi))
	}

	// Адаптация theta (MA коэффициент)
	if p.lastError != 0 {
		gradientTheta := -p.lastError * p.lastError
		p.theta -= learningRate * gradientTheta
		// Ограничение theta в разумных пределах
		p.theta = math.Max(-0.9, math.Min(0.9, p.theta))
	}
}

// GetStats возвращает текущую статистику наблюдений
func (p *ARIMAPredictor) GetStats() (mean, variance, trend float64) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Вычисление тренда (наклон линейной регрессии)
	trend = p.calculateTrend()

	return p.mean, p.variance, trend
}

// calculateTrend вычисляет тренд методом наименьших квадратов
func (p *ARIMAPredictor) calculateTrend() float64 {
	n := float64(len(p.window))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range p.window {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

// GetPredictedTPS возвращает прогнозируемую пропускную способность (TPS)
func (p *ARIMAPredictor) GetPredictedTPS(horizon int) float64 {
	forecast := p.Predict(horizon)
	if len(forecast) == 0 {
		return 0
	}

	// Возвращаем среднее из прогноза
	sum := 0.0
	for _, v := range forecast {
		sum += v
	}
	return sum / float64(len(forecast))
}

// GetLoadLevel возвращает уровень нагрузки (низкий/средний/высокий/пиковый)
func (p *ARIMAPredictor) GetLoadLevel() LoadLevel {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.mean < 100 {
		return LoadLevelLow
	} else if p.mean < 500 {
		return LoadLevelMedium
	} else if p.mean < 1000 {
		return LoadLevelHigh
	}
	return LoadLevelPeak
}

// LoadLevel - уровень нагрузки
type LoadLevel int

const (
	LoadLevelLow LoadLevel = iota
	LoadLevelMedium
	LoadLevelHigh
	LoadLevelPeak
)

func (l LoadLevel) String() string {
	switch l {
	case LoadLevelLow:
		return "LOW"
	case LoadLevelMedium:
		return "MEDIUM"
	case LoadLevelHigh:
		return "HIGH"
	case LoadLevelPeak:
		return "PEAK"
	default:
		return "UNKNOWN"
	}
}
