package service

type MemStorageSyncService struct {
	MemStorageService
	filename string
}

func NewMemStorageSyncService(service MemStorageService, filename string) *MemStorageSyncService {
	return &MemStorageSyncService{MemStorageService: service, filename: filename}
}
func (r *MemStorageSyncService) UpdateCount(name string, delta *int64) {
	r.repository.UpdateCount(name, delta)
	r.repository.LoadToFile(r.filename)
}

func (r *MemStorageSyncService) UpdateGauge(name string, value *float64) {
	r.repository.UpdateGauge(name, value)
	r.repository.LoadToFile(r.filename)
}
