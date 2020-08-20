/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package keycache

import (
	commLog "intel/isecl/lib/common/v3/log"
	"sync"
	"time"
)

var log = commLog.GetDefaultLogger()

// Key
type Key struct {
	ID      string
	Bytes   []byte
	Created time.Time
	Expired time.Time
}

// Cache is a mutex protected cache for quick storage and retrieval of keys by ImageUUID
// This implements an in-memory store only, and any data is effectively lost on application exit
type Cache struct {
	keys map[string]Key
	mtx  *sync.Mutex
}

// NewCache creates a new instance of a key cache
// It returns a pointer to the Cache struct
func NewCache() *Cache {
	log.Trace("keycache/keycache:NewCache() Entering")
	defer log.Trace("keycache/keycache:NewCache() Leaving")
	return &Cache{
		keys: make(map[string]Key),
		mtx:  &sync.Mutex{},
	}
}

// Get retrieves a key by its keyID
// It returns a byte slice containing the key data, as well as a bool that indicates
// if the key exists in the cache
func (c *Cache) Get(imageUUID string) (key Key, exists bool) {
	log.Trace("keycache/keycache:Get() Entering")
	defer log.Trace("keycache/keycache:Get() Leaving")
	c.mtx.Lock()
	defer c.mtx.Unlock()
	key, exists = c.keys[imageUUID]
	return
}

// Store persists a key in the cache by its keyID
func (c *Cache) Store(imageID string, key Key) {
	log.Trace("keycache/keycache:Store() Entering")
	defer log.Trace("keycache/keycache:Store() Leaving")
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.keys[imageID] = key
}

var global *Cache

func init() {
	log.Trace("keycache/keycache:init() Entering")
	defer log.Trace("keycache/keycache:init() Leaving")
	global = NewCache()
}

// Get retrieves a key by its imageID from the default global keycache
func Get(imageID string) (key Key, exists bool) {
	log.Trace("keycache/keycache:Get() Entering")
	defer log.Trace("keycache/keycache:Get() Leaving")
	return global.Get(imageID)
}

// Store persists a key by its keyID from the default global keycache
func Store(imageID string, key Key) {
	log.Trace("keycache/keycache:Store() Entering")
	defer log.Trace("keycache/keycache:Store() Leaving")
	global.Store(imageID, key)
}
