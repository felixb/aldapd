package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
	"os"
)

var opts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show more verbose logs"`
	Silent  bool   `short:"s" long:"silent" description:"Show critical messages only"`

	ListenAddr    string `short:"a" long:"address" default:"localhost" description:"Listen on this address"`
	ListenPort    uint32 `short:"p" long:"port" default:"389" description:"Listen on this port"`
	BaseDn        string `short:"b" long:"base-dn" default:"dc=felixb,dc=github,dc=com" description:"Present users and groups under this FDN"`
	AllowAnonBind bool   `long:"allow-anon-bind" description:"Allow bind with empty bind DN and password"`

	Files []string `short:"f" long:"file" required:"true" description:"Config file with user/group data"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	if opts.Silent {
		logging.SetLevel(logging.CRITICAL, "")
	} else if len(opts.Verbose) == 0 {
		logging.SetLevel(logging.WARNING, "")
	} else if len(opts.Verbose) == 1 {
		logging.SetLevel(logging.INFO, "")
	} else {
		logging.SetLevel(logging.DEBUG, "")
	}

	if backend, err := NewLocalFileBackend(opts.Files); err != nil {
		log.Panicf("error initializing backend: %s", err.Error())
	} else {
		c := &Config{
			listenAddr:    opts.ListenAddr,
			listenPort:    opts.ListenPort,
			allowAnonBind: opts.AllowAnonBind,
			baseDn:        opts.BaseDn,
			peopleDn:      fmt.Sprintf("ou=people,%s", opts.BaseDn),
			groupsDn:      fmt.Sprintf("ou=groups,%s", opts.BaseDn),
			backend:       backend,
		}

		s := NewServer(c)
		go s.signalHandler()
		if err := s.ListenAndServe(); err != nil {
			log.Errorf("error starting LDAP server: %s", err.Error())
		}
	}
}

