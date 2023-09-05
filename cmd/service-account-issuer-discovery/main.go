// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gardener/service-account-issuer-discovery/internal/app"
	"github.com/gardener/service-account-issuer-discovery/internal/version"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	showVersion          = flag.Bool("version", false, "Print the server version information.")
	kubeconfig           = flag.String("kubeconfig", "", "Path to the kubeconfig file. If not specified in cluster kubeconfig will be used.")
	hostname             = flag.String("hostname", "", "Hostname to serve the public keys on.")
	certFile             = flag.String("cert-file", "", "Path to certificate file.")
	keyFile              = flag.String("key-file", "", "Path to key file.")
	port                 = flag.Int("port", 0, "Port to start the server on. Defaults to 443 when tls is enabled and 10443 when it is not.")
	cacheRefreshInterval = flag.Int64("cache-refresh-interval", 10, "The number of seconds between each response cache refresh. Defaults to 10.")
	cachedObjectTTL      = flag.Int64("cached-object-ttl", 30, "The number of seconds to retain a cached response before discarding it from cache on refresh. Defaults to 30.")
)

func main() {
	flag.Parse()

	if *showVersion {
		info, err := version.BuildInfo()
		if err != nil {
			log.Fatal(err)
		}

		jsonInfo, err := json.Marshal(info)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonInfo))
		os.Exit(0)
	}

	restConfig, err := getRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
	restConfig.Timeout = time.Second * 10

	handler, err := app.NewHandlersSet(&app.HandlersConfig{
		Hostname:                      hostname,
		RESTConfig:                    restConfig,
		CacheRefreshIntervalInSeconds: cacheRefreshInterval,
		CachedObjectValidityInSeconds: cachedObjectTTL,
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", handler.Config)
	mux.HandleFunc("/jwks", handler.JWKS)
	mux.HandleFunc("/healthz", handler.Healthz)

	if certFile != nil && *certFile != "" && keyFile != nil && *keyFile != "" {
		certGetter, err := app.NewCertificateGetter(*certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}

		tlsConf := &tls.Config{
			GetCertificate: certGetter.Certificate(),
			MinVersion:     tls.VersionTLS13,
		}

		if port == nil || *port == 0 {
			*port = 443
		}

		server := &http.Server{
			Addr:         fmt.Sprintf(":%v", *port),
			Handler:      mux,
			TLSConfig:    tlsConf,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		log.Printf("Starting server on port %v ...", *port)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	} else {
		if port == nil || *port == 0 {
			*port = 10443
		}

		server := &http.Server{
			Addr:         fmt.Sprintf(":%v", *port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		log.Printf("Starting server on port %v ...", *port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}

func getRESTConfig() (*rest.Config, error) {
	var cfg *rest.Config
	var kubeconfigFilePath string
	if kubeconfig != nil && *kubeconfig != "" {
		kubeconfigFilePath = *kubeconfig
	} else if kubecfg, ok := os.LookupEnv("KUBECONFIG"); ok {
		kubeconfigFilePath = kubecfg
	}

	if len(kubeconfigFilePath) != 0 {
		kubeconfigBytes, err := os.ReadFile(kubeconfigFilePath)
		if err != nil {
			return nil, err
		}

		cfg, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
