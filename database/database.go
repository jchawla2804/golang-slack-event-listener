package database

import (
	"time"

	"github.com/patrickmn/go-cache"
)

func CreateCache() *cache.Cache {
	c := cache.New(1800*time.Second, 1800*time.Second)
	return c
}
