package service

type Cacheable interface {
	Warmup()
}
