package repo

type KVRepo[T any] interface {
	Set(key string, value *T, ttlSeconds int64) error
	Save(key string, ttlSeconds int64, saveFunc func(*T) error) error
	Touch(key string, ttlSeconds int64) error
	Delete(key string) error
	Get(key string) (*T, error)
	GetAll() ([]*T, error)
	GetAllKeys() ([]string, error)
	KeyExists(key string) (bool, error)
	LastModifiedSeconds(key string) (int64, error)
}
