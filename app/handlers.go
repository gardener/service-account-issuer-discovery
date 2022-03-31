// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/singleflight"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	notAllowedTempl = "%s method is not allowed"
	genericError    = "Could not retrieve the requested configuration"
	configKey       = "config"
	jwksKey         = "jwks"
)

type handlersSet struct {
	hostname     string
	clientset    *kubernetes.Clientset
	logger       *log.Logger
	cacher       Cacher
	requestGroup singleflight.Group
}

// ServiceAccountDiscoveryHandler is a set of handler functions that handle OpenID discovery documents
type ServiceAccountDiscoveryHandler interface {
	// Config is a handler that serves the /.well-known/openid-configuration endpoint of a Kubernetes cluster
	Config(w http.ResponseWriter, r *http.Request)
	// JWKS is a handler that serves the public keys used to verify service account tokens of a Kubernetes cluster
	JWKS(w http.ResponseWriter, r *http.Request)
	// Healthz is a handler used to determine the state of the proxy
	Healthz(w http.ResponseWriter, r *http.Request)
}

type HandlersConfig struct {
	Hostname                      *string
	RESTConfig                    *rest.Config
	CacheRefreshIntervalInSeconds *int64
	CachedObjectValidityInSeconds *int64
}

func NewHandlersSet(config *HandlersConfig) (ServiceAccountDiscoveryHandler, error) {
	if config.Hostname == nil || *config.Hostname == "" {
		return nil, fmt.Errorf("hostname should not be empty")
	}

	clientset, err := kubernetes.NewForConfig(config.RESTConfig)
	if err != nil {
		return nil, err
	}

	cacheRefresh := time.Second * 10
	if config.CacheRefreshIntervalInSeconds != nil && *config.CacheRefreshIntervalInSeconds > 0 {
		parsed, err := time.ParseDuration(fmt.Sprintf("%vs", *config.CacheRefreshIntervalInSeconds))
		if err != nil {
			return nil, err
		}
		cacheRefresh = parsed
	}

	var cacheTTL int64 = 30
	if config.CachedObjectValidityInSeconds != nil && *config.CachedObjectValidityInSeconds > 0 {
		cacheTTL = *config.CachedObjectValidityInSeconds
	}

	cacher, err := NewCacher(cacheRefresh, cacheTTL)
	if err != nil {
		return nil, err
	}

	return &handlersSet{
		hostname:  *config.Hostname,
		clientset: clientset,
		logger:    log.Default(),
		cacher:    cacher,
	}, nil
}

func (s *handlersSet) Config(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf(notAllowedTempl, r.Method), http.StatusMethodNotAllowed)
		return
	}

	cached := s.cacher.Get(configKey)
	if cached != nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	defer s.requestGroup.Forget(configKey)
	v, err, _ := s.requestGroup.Do(configKey, func() (interface{}, error) {
		// If we miss the cache then we make the actual call to the apiserver but limit it to a single request
		oidcConfig, err := getOIDConfig(r.Context(), s.clientset.RESTClient())
		if err != nil {
			return nil, err
		}

		issuerURL, err := url.Parse(oidcConfig.Issuer)
		if err != nil {
			return nil, err
		}

		if issuerURL.Host != s.hostname {
			return nil, fmt.Errorf("apiserver's issuer does not match the configured hostname")
		}

		oidcConfig.Issuer = fmt.Sprintf("https://%s", s.hostname)
		oidcConfig.JWKSURL = fmt.Sprintf("https://%s/jwks", s.hostname)

		bytes, err := json.Marshal(oidcConfig)
		if err != nil {
			return nil, err
		}

		return bytes, nil
	})
	if err != nil {
		http.Error(w, genericError, http.StatusInternalServerError)
		s.logger.Println(err)
		return
	}

	// Conversion should be safe
	bytes := v.([]byte)
	s.cacher.Update(configKey, bytes)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (s *handlersSet) JWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf(notAllowedTempl, r.Method), http.StatusMethodNotAllowed)
		return
	}

	cached := s.cacher.Get(jwksKey)
	if cached != nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	defer s.requestGroup.Forget(jwksKey)
	v, err, _ := s.requestGroup.Do(jwksKey, func() (interface{}, error) {
		// If we miss the cache then we make the actual call to the apiserver but limit it to a single request
		oidcConfig, err := getOIDConfig(r.Context(), s.clientset.RESTClient())
		if err != nil {
			return nil, err
		}

		jwksURL, err := url.Parse(oidcConfig.JWKSURL)
		if err != nil {
			return nil, err
		}

		jwksBytes, err := getJWKS(r.Context(), s.clientset.RESTClient(), jwksURL.Path)
		if err != nil {
			return nil, err
		}
		return jwksBytes, nil
	})
	if err != nil {
		http.Error(w, genericError, http.StatusInternalServerError)
		s.logger.Println(err)
		return
	}

	// Conversion should be safe
	jwksBytes := v.([]byte)
	s.cacher.Update(jwksKey, jwksBytes)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jwksBytes)
}

func (s *handlersSet) Healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf(notAllowedTempl, r.Method), http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte("ok"))
}

type oidConfig struct {
	Issuer                 string   `json:"issuer,omitempty"`
	JWKSURL                string   `json:"jwks_uri,omitempty"`
	ResponseTypesSupported []string `json:"response_types_supported,omitempty"`
	SubjectTypesSupported  []string `json:"subject_types_supported,omitempty"`
	SigningAlgsSupported   []string `json:"id_token_signing_alg_values_supported,omitempty"`
}

func getOIDConfig(ctx context.Context, client rest.Interface) (*oidConfig, error) {
	oidReq := client.Get()
	oidReq.RequestURI("/.well-known/openid-configuration")
	respBytes, err := oidReq.DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	oidConfig := &oidConfig{}
	err = json.Unmarshal(respBytes, oidConfig)
	if err != nil {
		return nil, err
	}
	return oidConfig, nil
}

func getJWKS(ctx context.Context, client rest.Interface, relativeUri string) ([]byte, error) {
	jwksReq := client.Get()
	jwksReq.RequestURI(relativeUri)
	return jwksReq.DoRaw(ctx)
}
