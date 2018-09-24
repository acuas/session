package memory

import (
	"sync"

	"github.com/fasthttp/session"
)

// Config session memory configuration
type Config struct{}

// Provider provider struct
type Provider struct {
	config      *Config
	values      *session.Dict
	maxLifeTime int64

	lock sync.RWMutex
}

// Store memory store
type Store struct {
	session.Store

	lastActiveTime int64
	lock           sync.RWMutex
}