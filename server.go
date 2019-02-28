package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark-rushakoff/ldapserver"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("aldapd")
)

type Config struct {
	listenAddr    string
	listenPort    uint32
	allowAnonBind bool
	baseDn        string
	peopleDn      string
	groupsDn      string
	backend       Backender
}

type Server struct {
	config     *Config
	backend    Backender
	ldapServer *ldapserver.Server
}

func NewServer(config *Config) *Server {
	s := &Server{
		config:     config,
		backend:    config.backend,
		ldapServer: ldapserver.NewServer(),
	}

	s.ldapServer.Bind = s.bind
	s.ldapServer.Search = s.search

	return s

}

func StartServer(config *Config) *Server {
	s := NewServer(config)
	go s.ListenAndServe()
	return s
}

func (s *Server) ListenAndServe() error {
	listen := fmt.Sprintf("%s:%d", s.config.listenAddr, s.config.listenPort)
	log.Infof("starting example LDAP server on %s with base dn %s", listen, s.config.baseDn)
	if err := s.ldapServer.ListenAndServe(listen); err != nil {
		return err
	}
	return nil
}

func (s *Server) Reload() {
	if err := s.backend.Reload(); err != nil {
		log.Errorf("error reloading backend data: %s", err.Error())
	}
}

func (s *Server) Close() {
	log.Info("shutting down LDAP server")
	s.ldapServer.Close()
}

func (s *Server) signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	signal.Notify(c, syscall.SIGUSR1)
	signal.Notify(c, syscall.SIGUSR2)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)

	for sig := range c {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			s.Close()
		} else {
			s.Reload()
		}
	}
}
