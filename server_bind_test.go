package main

import (
	"fmt"
	"testing"

	"github.com/mark-rushakoff/ldapserver"
	"github.com/stretchr/testify/assert"
)

func TestServer_bind(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return username == password, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.NoError(t, conn.Bind("cn=foo,ou=people,ou=test,dc=example,dc=com", "foo"))
	assert.NoError(t, conn.Bind("cn=FOo,ou=people,ou=test,dc=example,dc=com", "foo"))
	assert.Error(t, conn.Bind("cn=foo,ou=people,ou=test,dc=example,dc=com", "bar"))
}

func TestServer_bind_anonymous(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		assert.Fail(t, "bind() should not have been called")
		return false, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)

	s.config.allowAnonBind = true
	assert.NoError(t, conn.Bind("", ""))

	s.config.allowAnonBind = false
	assert.Error(t, conn.Bind("", ""))
}

func TestServer_bind_error(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return true, fmt.Errorf("some error")
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.Error(t, conn.Bind("foo", "foo"))
}
