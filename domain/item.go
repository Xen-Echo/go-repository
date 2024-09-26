package domain

type Item[T any] struct {
	Key               string `json:"key"`
	Value             *T     `json:"value"`
	TTLSeconds        int64  `json:"ttl_seconds"`
	ModifiedAtSeconds int64  `json:"modified_at_seconds"`
}
