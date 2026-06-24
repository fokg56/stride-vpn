package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type StoredConfig struct {
	ID     string `json:"id"`
	Link   string `json:"link"`
	Remark string `json:"remark"`
}

type Store struct {
	path    string
	mu      sync.RWMutex
	configs []StoredConfig
}

func NewStore(path string) *Store {
	if path == "" {
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		path = filepath.Join(dir, "configs.json")
	}
	return &Store{path: path}
}

func (s *Store) Load() ([]StoredConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.configs = []StoredConfig{}
			return s.configs, nil
		}
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if err := json.Unmarshal(data, &s.configs); err != nil {
		s.configs = []StoredConfig{}
		return s.configs, fmt.Errorf("ошибка парсинга: %w", err)
	}

	return s.configs, nil
}

func (s *Store) Save(configs []StoredConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs = configs
	data, err := json.MarshalIndent(s.configs, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	return nil
}

func (s *Store) GetPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}
