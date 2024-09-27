package repo

import (
	"github.com/xen-echo/go-repository/datasource"
	"github.com/xen-echo/go-repository/service"
)

type disk[T any] struct {
	name string
	ds   datasource.SFPKDiskDS[T]
}

func NewDiskKVRepo[T any](name string) KVRepo[T] {
	return &disk[T]{name: name, ds: datasource.NewSFPKDiskDS[T](name)}
}

func NewDiskKVRepoWithEncryption[T any](name string, encryptionService service.EncryptionService) KVRepo[T] {
	return &disk[T]{name: name, ds: datasource.NewSFPKDiskDSWithEncryption[T](name, encryptionService)}
}

func (d *disk[T]) Set(key string, value *T, ttlSeconds int64) error {
	df, err := d.ds.GetDataFile(key)
	if err != nil {
		return err
	}
	defer df.Unlock()
	df.Item.Value = value
	df.Item.TTLSeconds = ttlSeconds
	return d.ds.SaveDataFile(df)
}

func (d *disk[T]) Touch(key string, ttlSeconds int64) error {
	df, err := d.ds.GetDataFile(key)
	if err != nil {
		return err
	}
	defer df.Unlock()
	df.Item.TTLSeconds = ttlSeconds
	return d.ds.SaveDataFile(df)
}

func (d *disk[T]) Delete(key string) error {
	return d.ds.DeleteDataFile(key)
}

func (d *disk[T]) Get(key string) (*T, error) {
	df, err := d.ds.GetDataFile(key)
	if err != nil {
		return nil, err
	}
	defer df.Unlock()
	return df.Item.Value, nil
}

func (d *disk[T]) GetAll() ([]*T, error) {
	dfs, err := d.ds.GetAllDataFiles()
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, df := range dfs {
			df.Unlock()
		}
	}()
	var values []*T
	for _, df := range dfs {
		values = append(values, df.Item.Value)
	}
	return values, nil
}

func (d *disk[T]) KeyExists(key string) (bool, error) {
	return d.ds.ExistsDataFile(key)
}
