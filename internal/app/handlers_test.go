// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"testing"

	"github.com/gardener/service-account-issuer-discovery/internal/app"

	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

func TestInvalidHandlerConfigs(t *testing.T) {
	tests := []struct {
		name          string
		config        *app.HandlersConfig
		expectedError string
	}{
		{
			name:          "hostname is nil",
			config:        &app.HandlersConfig{},
			expectedError: "hostname should not be empty",
		},
		{
			name:          "hostname is empty string",
			config:        &app.HandlersConfig{},
			expectedError: "hostname should not be empty",
		},
		{
			name: "refresh interval is greated than single cached object ttl",
			config: &app.HandlersConfig{
				Hostname:                      pointer.String("test"),
				RESTConfig:                    &rest.Config{},
				CacheRefreshIntervalInSeconds: pointer.Int64(4),
				CachedObjectValidityInSeconds: pointer.Int64(3),
			},
			expectedError: "the refresh interval of 4 seconds should not be greater than the cached object validity duration seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := app.NewHandlersSet(tt.config)
			if h != nil {
				t.Errorf("Expected not to return handler but got: %v", h)
			}
			if err.Error() != tt.expectedError {
				t.Errorf(`Expected error "%s", but got "%s"`, tt.expectedError, err.Error())
			}
		})
	}
}
