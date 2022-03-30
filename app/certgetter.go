// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"crypto/tls"
	"log"
	"sync"
	"time"
)

type reloader struct {
	mutex       sync.RWMutex
	certificate *tls.Certificate
	certFile    string
	keyFile     string
	logger      *log.Logger
}

type CertificateGetter interface {
	Certificate() func(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

func NewCertificateGetter(certFile, keyFile string) (CertificateGetter, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	reloader := &reloader{
		mutex:       sync.RWMutex{},
		certificate: &cert,
		certFile:    certFile,
		keyFile:     keyFile,
		logger:      log.Default(),
	}

	go func() {
		for range time.Tick(time.Hour * 12) {
			if err := reloader.reload(); err != nil {
				reloader.logger.Println(err)
			}
		}
	}()

	return reloader, nil
}

func (r *reloader) Certificate() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		r.mutex.RLock()
		defer r.mutex.RUnlock()
		return r.certificate, nil
	}
}

func (r *reloader) reload() error {
	newCert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		return err
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.certificate = &newCert
	r.logger.Println("certificate successfully renewed")
	return nil
}
