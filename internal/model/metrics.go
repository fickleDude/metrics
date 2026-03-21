package models

import (
	"fmt"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) GetValue() string {
	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return "значение не задано"
		}
		return strconv.FormatInt(*m.Delta, 10)
	case Gauge:
		if m.Value == nil {
			return "значение не задано"
		}
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	default:
		return "значение не задано"
	}
}

func (m *Metrics) SetValue(value interface{}) error {
	switch m.MType {
	case "gauge":
		if v, ok := value.(*float64); ok {
			m.Value = v
		} else if v, ok := value.(*uint64); ok {
			convert := float64(*v)
			m.Value = &convert
		} else if v, ok := value.(*uint32); ok {
			convert := float64(*v)
			m.Value = &convert
		} else if v, ok := value.(float64); ok {
			m.Value = &v
		} else {
			return fmt.Errorf("не получилось преобразовать тип interface{} в тип float64")
		}
	case "counter":
		if v, ok := value.(*int64); ok {
			m.Delta = v
		} else {
			return fmt.Errorf("не получилось преобразовать тип interface{} в тип int64")
		}
	default:
		return fmt.Errorf("неизвестный тип метрики %s", m.MType)

	}
	return nil
}
