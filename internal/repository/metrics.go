package repository

type MetricsStorage struct {
	count map[string]int64
	gauge map[string]float64
}

type MemStorageRepository struct {
	storage MetricsStorage
}

func (r *MemStorageRepository) WriteCount(name string, value int64) error {
	return nil
}

func (r *MemStorageRepository) WriteGauge(name string, value int64) error {
	return nil
}
