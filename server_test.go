package main

import (
	"fmt"
	"github.com/op/go-logging"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	listenAddr             = "localhost"
	waitBeforeRunningTests = time.Millisecond * 50
)

type TestBackend struct {
	bindFunc   func(username, password string) (bool, error)
	usersFunc  func(filterKey, filterValue string) ([]User, error)
	groupsFunc func(filterKey, filterValue string) ([]Group, error)
}

func (tb *TestBackend) Check(username, password string) (bool, error) {
	return tb.bindFunc(username, password)
}

func (tb *TestBackend) Users(filterKey, filterValue string) ([]User, error) {
	return tb.usersFunc(filterKey, filterValue)
}

func (tb *TestBackend) Groups(filterKey, filterValue string) ([]Group, error) {
	return tb.groupsFunc(filterKey, filterValue)
}

func (tb *TestBackend) Reload() error {
	return nil
}

func startTestServer(t *testing.T) (*Server, *TestBackend, string) {
	logging.SetLevel(logging.DEBUG, "")

	listenPort := uint32(rand.Int31n(60000) + 1024)
	listen := fmt.Sprintf("%s:%d", listenAddr, listenPort)

	tb := &TestBackend{}
	config := &Config{
		listenAddr:    listenAddr,
		listenPort:    listenPort,
		allowAnonBind: false,
		baseDn:        "ou=test,dc=example,dc=com",
		peopleDn:      "ou=people,ou=test,dc=example,dc=com",
		groupsDn:      "ou=groups,ou=test,dc=example,dc=com",
		backend:       tb}

	s := StartServer(config)
	assert.NotNil(t, s)

	time.Sleep(waitBeforeRunningTests)

	return s, tb, listen
}
