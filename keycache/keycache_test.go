/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package keycache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var t1 = time.Now()
var expTimeDelta, _ = time.ParseDuration("5m")
var t2 = t1.Add(expTimeDelta)

func TestGetAndStore(t *testing.T) {
	log.Trace("keycache/keycache_test:TestGetAndStore() Entering")
	defer log.Trace("keycache/keycache_test:TestGetAndStore() Leaving")
	assert := assert.New(t)
	cache := NewCache()

	key := Key{"keyid", []byte{0, 1, 2, 3}, t1, t2}
	cache.Store("foobar", key)
	actual, exists := cache.Get("foobar")
	assert.True(exists)
	assert.Equal(key, actual)
}

func TestGetNone(t *testing.T) {
	log.Trace("keycache/keycache_test:TestGetNone() Entering")
	defer log.Trace("keycache/keycache_test:TestGetNone() Leaving")
	assert := assert.New(t)
	cache := NewCache()
	actual, exists := cache.Get("foobar")
	assert.False(exists)
	assert.Zero(actual)
}

func TestOverwrite(t *testing.T) {
	log.Trace("keycache/keycache_test:TestOverwrite() Entering")
	defer log.Trace("keycache/keycache_test:TestOverwrite() Leaving ")
	assert := assert.New(t)
	cache := NewCache()
	key1 := Key{"foo", []byte{0, 1, 2, 3}, t1, t2}
	key2 := Key{"bar", []byte{4, 5, 6, 7}, t1, t2}
	cache.Store("foobar", key1)
	cache.Store("foobar", key2)
	actual, exists := cache.Get("foobar")
	assert.True(exists)
	assert.Equal(key2, actual)
}
