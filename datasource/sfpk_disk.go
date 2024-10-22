package datasource

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/xen-echo/go-repository/domain"
	"github.com/xen-echo/go-repository/service"
)

// SINGLE FILE PER KEY DISK DATAFILE

type SFPKDiskDatafile[T any] struct {
	FilePath string
	Item     domain.Item[T]
	mu       *sync.Mutex
}

func (d *SFPKDiskDatafile[T]) Unlock() {
	d.mu.Unlock()
}

type SFPKDiskDS[T any] interface {
	GetDataFile(name string) (*SFPKDiskDatafile[T], error)
	GetAllDataFiles() ([]*SFPKDiskDatafile[T], error)
	SaveDataFile(dataFile *SFPKDiskDatafile[T]) error
	DeleteDataFile(name string) error
	ExistsDataFile(name string) (bool, error)
	Wipe() error
}

type sfpk[T any] struct {
	root              string
	name              string
	muMap             sync.Map
	encryptionService service.EncryptionService
}

func NewSFPKDiskDS[T any](name string) SFPKDiskDS[T] {
	root := os.Getenv("SFPK_ROOT")
	if root == "" {
		root = "./sfpk"
	}
	return &sfpk[T]{root: root, name: name, muMap: sync.Map{}}
}

func NewSFPKDiskDSWithEncryption[T any](name string, encryptionService service.EncryptionService) SFPKDiskDS[T] {
	ds := NewSFPKDiskDS[T](name)
	ds.(*sfpk[T]).encryptionService = encryptionService
	return ds
}

func (s *sfpk[T]) getFileDir() (string, error) {
	p := path.Join(s.root, s.name)
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (s *sfpk[T]) getFilePath(filename string) (string, error) {
	d, err := s.getFileDir()
	if err != nil {
		return "", err
	}
	extension := "json"
	if s.encryptionService != nil {
		extension = "enc"
	}
	return path.Join(d, fmt.Sprintf("%s.%s", filename, extension)), nil
}

func (s *sfpk[T]) getFileMutex(name string) *sync.Mutex {
	mu, _ := s.muMap.LoadOrStore(name, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

func (s *sfpk[T]) GetDataFile(name string) (*SFPKDiskDatafile[T], error) {
	filePath, err := s.getFilePath(name)
	if err != nil {
		return nil, err
	}

	mu := s.getFileMutex(name)
	mu.Lock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := os.WriteFile(filePath, []byte("{}"), os.ModePerm)
		if err != nil {
			mu.Unlock()
			return nil, err
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		mu.Unlock()
		return nil, err
	}

	if s.encryptionService != nil && string(data) != "{}" {
		data, err = s.encryptionService.Decrypt(data)
		if err != nil {
			mu.Unlock()
			return nil, err
		}
	}

	item := domain.Item[T]{}
	err = json.Unmarshal(data, &item)
	if err != nil {
		mu.Unlock()
		return nil, err
	}
	item.Key = name

	if item.TTLSeconds > 0 && item.ModifiedAtSeconds+item.TTLSeconds < time.Now().Unix() {
		err := os.Remove(filePath)
		mu.Unlock()
		if err != nil {
			return nil, err
		}
		return s.GetDataFile(name)
	}

	return &SFPKDiskDatafile[T]{FilePath: filePath, Item: item, mu: mu}, nil
}

func (s *sfpk[T]) GetAllDataFiles() ([]*SFPKDiskDatafile[T], error) {
	fileDir, err := s.getFileDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, err
	}

	dataFiles := make([]*SFPKDiskDatafile[T], 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Get the name and strip the extension
		name := file.Name()
		parts := strings.Split(name, ".")
		if len(parts) > 1 {
			name = strings.Join(parts[:len(parts)-1], ".")
		}

		dataFile, err := s.GetDataFile(name)
		if err != nil {
			return nil, err
		}

		dataFiles = append(dataFiles, dataFile)
	}

	return dataFiles, nil
}

func (s *sfpk[T]) SaveDataFile(dataFile *SFPKDiskDatafile[T]) error {
	dataFile.Item.ModifiedAtSeconds = time.Now().Unix()

	data, err := json.Marshal(dataFile.Item)
	if err != nil {
		return err
	}

	if s.encryptionService != nil {
		data, err = s.encryptionService.Encrypt(data)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(dataFile.FilePath, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (s *sfpk[T]) DeleteDataFile(name string) error {
	filePath, err := s.getFilePath(name)
	if err != nil {
		return err
	}

	err = os.Remove(filePath)
	if err != nil {
		return err
	}

	s.muMap.Delete(name)

	return nil
}

func (s *sfpk[T]) ExistsDataFile(name string) (bool, error) {
	filePath, err := s.getFilePath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (s *sfpk[T]) Wipe() error {
	fileDir, err := s.getFileDir()
	if err != nil {
		return err
	}
	err = os.RemoveAll(fileDir)
	if err != nil {
		return err
	}
	return nil
}
