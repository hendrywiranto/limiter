package limiter

type Limiter struct {
	adapter Adapter
}

func New(adapter Adapter) *Limiter {
	return &Limiter{
		adapter: adapter,
	}
}
