// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"testing"
	"time"

	"github.com/gardener/service-account-issuer-discovery/pkg/cache"
)

func TestEmptyCache(t *testing.T) {
	c, err := cache.NewCache(time.Second*2, 2)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	actual := c.Get("something")
	if actual != nil {
		t.Errorf("Result was, got: %v, want: %v.", actual, nil)
	}
}

func TestRetrieveFromCache(t *testing.T) {
	expected := "hello"
	c, err := cache.NewCache(time.Second*2, 2)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	c.Update("something", []byte(expected))

	actual := string(c.Get("something"))
	if actual != expected {
		t.Errorf("Result was, got: %v, want: %v.", actual, expected)
	}
}

func TestRetrieveFromCacheMultipleObjects(t *testing.T) {
	expected := "hello"
	c, err := cache.NewCache(time.Second*2, 2)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	c.Update("something", []byte(expected))
	c.Update("something_else", []byte("unexpected"))

	actual := string(c.Get("something"))
	if actual != expected {
		t.Errorf("Result was, got: %v, want: %v.", actual, expected)
	}
}

func TestExpiredResultCache(t *testing.T) {
	c, err := cache.NewCache(time.Second*2, 2)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	c.Update("something", []byte("hello"))

	time.Sleep(time.Second * 4)
	actual := c.Get("something")
	if actual != nil {
		t.Errorf("Result was, got: %v, want: %v.", actual, nil)
	}
}

func TestStopRefreshingCache(t *testing.T) {
	expected := "hello"
	c, err := cache.NewCache(time.Second*2, 2)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	c.Update("something", []byte(expected))

	c.StopRefresh()
	time.Sleep(time.Second * 4)
	actual := string(c.Get("something"))
	if actual != expected {
		t.Errorf("Result was, got: %v, want: %v.", actual, expected)
	}
}

func TestInvalidCacheConfig(t *testing.T) {
	expectedError := "the refresh interval of 3 seconds should not be greater than the cached object validity duration seconds"
	c, err := cache.NewCache(time.Second*3, 2)
	if c != nil {
		t.Errorf("Expected not to return cacher but got: %v", c)
	}
	if err.Error() != expectedError {
		t.Errorf(`Expected error "%s", but got "%s"`, expectedError, err.Error())
	}
}
