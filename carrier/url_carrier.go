package carrier

import (
	"net/url"
	"sync"

	"go.opentelemetry.io/otel/propagation"
)

// NewURLCarrier returns a carrier adapts url.URL to satisfy the propagation.TextMapCarrier interface.
func NewURLCarrier(u *url.URL) propagation.TextMapCarrier {
	return &urlCarrier{URL: u}
}

type urlCarrier struct {
	*url.URL
	sync.RWMutex
}

func (uc *urlCarrier) Get(key string) string {
	uc.RLock()
	defer uc.RUnlock()
	return uc.Query().Get(key)
}

func (uc *urlCarrier) Set(key, value string) {
	uc.Lock()
	defer uc.Unlock()
	qs := uc.Query()
	qs.Set(key, value)
	uc.RawQuery = qs.Encode()
}

func (uc *urlCarrier) Keys() []string {
	uc.RLock()
	defer uc.RUnlock()
	qs := uc.Query()
	keys := make([]string, 0, len(qs))
	for k := range qs {
		keys = append(keys, k)
	}
	return keys
}
