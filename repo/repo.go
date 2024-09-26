package repo

type KVRepo[T any] interface {
	Set(key string, value *T, ttlSeconds int64) error
	Touch(key string, ttlSeconds int64) error
	Delete(key string) error
	Get(key string) (*T, error)
	GetAll() ([]*T, error)
}
