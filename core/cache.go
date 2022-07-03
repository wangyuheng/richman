package core

type Cache interface {
	Store(key string, val interface{})
	Load(key string) (value interface{}, ok bool)
}
