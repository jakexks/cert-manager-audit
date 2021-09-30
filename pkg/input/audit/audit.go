// Package audit is an audit adapter that receives events as a Kubernetes API Audit webhook
package audit

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"

	"sandbox.jakexks.dev/cert-manager-audit/pkg/input"
	"sandbox.jakexks.dev/cert-manager-audit/pkg/process"

)

type Audit struct {
	log         logr.Logger
	tlsConfig   *tls.Config
	processFunc process.Func
	config
}

type config struct {
	ListenAddr  string `json:"listen_addr,omitempty"`
	TLSCertFile string `json:"tls_cert_file,omitempty"`
	TLSKeyFile  string `json:"tls_key_file,omitempty"`

	// Auth options
	RequireClientAuth bool   `json:"require_client_auth"`
	UseSystemRoots    bool   `json:"use_system_roots"`
	CAFile            string `json:"client_auth_ca_file,omitempty"`
}

func (a *Audit) Start(ctx context.Context) error {
	a.log.Info("Starting Audit Log Webhook receiver")
	server := &http.Server{
		Addr:              a.ListenAddr,
		TLSConfig:         a.tlsConfig,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handler)
	server.Handler = mux

	go func(s *http.Server) {
		var err error
		a.log.Info("Starting HTTP server", "config", a.config)
		if len(a.TLSCertFile) > 0 {
			err = s.ListenAndServeTLS("", "")
		} else {
			err = s.ListenAndServe()
		}
		a.log.Error(err, "Audit webhook receiver shutting down")
	}(server)

	go func(ctx context.Context, s *http.Server) {
		<-ctx.Done()
		a.log.Info("Shutting down audit webhook receiver")
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(c); err != nil {
			a.log.Error(err, "Error shutting down HTTP server")
		}
	}(ctx, server)

	return nil
}

func (a *Audit) Stop(ctx context.Context) error {
	return nil
}

func (a *Audit) Setup(baseLogger logr.Logger, processFunc process.Func, inputConfig input.Config) error {
	a.log = baseLogger
	a.processFunc = processFunc
	a.defaults()

	cfg := new(config)
	err := yaml.Unmarshal(inputConfig, cfg)
	if err != nil {
		return err
	}

	if len(cfg.ListenAddr) > 0 {
		a.ListenAddr = cfg.ListenAddr
	}

	a.tlsConfig = &tls.Config{}
	if len(cfg.TLSCertFile) > 0 {
		tlsCert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("while loading cert and key %s, %s: %w",
				cfg.TLSCertFile,
				cfg.TLSKeyFile,
				err)
		}
		a.tlsConfig.Certificates = append(a.tlsConfig.Certificates, tlsCert)
	}

	if cfg.RequireClientAuth {
		a.tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

		var pool *x509.CertPool
		if cfg.UseSystemRoots {
			pool, err = x509.SystemCertPool()
			if err != nil {
				return fmt.Errorf("could not get system cert pool: %w", err)
			}
		} else {
			pool = x509.NewCertPool()
		}
		if len(cfg.CAFile) > 0 {
			pemData, err := ioutil.ReadFile(cfg.CAFile)
			if err != nil {
				return fmt.Errorf("while reading CAFile %s: %w", cfg.CAFile, err)
			}
			pool.AppendCertsFromPEM(pemData)
		}

		a.tlsConfig.ClientCAs = pool
	}

	return nil
}

func (a *Audit) defaults() {
	a.ListenAddr = ":8080"
}

type auditAdapter struct{}

func (auditAdapter) New() input.Input {
	return &Audit{}
}

func (auditAdapter) Name() string {
	return "audit"
}

func init() {
	input.Register(auditAdapter{})
}
