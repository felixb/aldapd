package main

import (
	"fmt"
	"testing"

	"github.com/mark-rushakoff/ldapserver"
	"github.com/stretchr/testify/assert"
)

func TestServer_search_users(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return true, nil
	}
	tb.usersFunc = func(filterKey, filterValue string) ([]User, error) {
		assert.Empty(t, filterKey)
		assert.Empty(t, filterValue)
		return []User{
			newTestUser("u1"),
			newTestUser("u2"),
		}, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.NoError(t, conn.Bind("cn=foo,ou=test,dc=example,dc=com", "foo"))
	r, err := conn.Search(&ldapserver.SearchRequest{
		BaseDN: "ou=people,ou=test,dc=example,dc=com",
		Filter: "(objectClass=*)",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 2, len(r.Entries))
	assertUser(t, "u1", r.Entries[0])
	assertUser(t, "u2", r.Entries[1])
}

func TestServer_search_user(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return true, nil
	}
	tb.usersFunc = func(filterKey, filterValue string) ([]User, error) {
		assert.Equal(t, "cn", filterKey)
		assert.Equal(t, "u1", filterValue)
		return []User{
			newTestUser("u1"),
		}, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.NoError(t, conn.Bind("cn=foo,ou=test,dc=example,dc=com", "foo"))
	r, err := conn.Search(&ldapserver.SearchRequest{
		BaseDN: "ou=people,ou=test,dc=example,dc=com",
		Filter: "(cn=u1)",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 1, len(r.Entries))
	assertUser(t, "u1", r.Entries[0])
}

func TestServer_search_groups(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return true, nil
	}
	tb.groupsFunc = func(filterKey, filterValue string) ([]Group, error) {
		assert.Empty(t, filterKey)
		assert.Empty(t, filterValue)
		return []Group{
			newTestGroup("g1"),
			newTestGroup("g2"),
		}, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.NoError(t, conn.Bind("cn=foo,ou=test,dc=example,dc=com", "foo"))
	r, err := conn.Search(&ldapserver.SearchRequest{
		BaseDN: "ou=groups,ou=test,dc=example,dc=com",
		Filter: "(objectClass=*)",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 2, len(r.Entries))
	assertGroup(t, "g1", r.Entries[0])
	assertGroup(t, "g2", r.Entries[1])
}

func TestServer_search_group(t *testing.T) {
	s, tb, listen := startTestServer(t)
	defer s.Close()

	tb.bindFunc = func(username, password string) (bool, error) {
		return true, nil
	}
	tb.groupsFunc = func(filterKey, filterValue string) ([]Group, error) {
		assert.Equal(t, "cn", filterKey)
		assert.Equal(t, "g1", filterValue)
		return []Group{
			newTestGroup("g1"),
		}, nil
	}

	conn, err := ldapserver.Dial("tcp", listen)
	assert.NotNil(t, conn)
	assert.NoError(t, err)
	assert.NoError(t, conn.Bind("cn=foo,ou=test,dc=example,dc=com", "foo"))
	r, err := conn.Search(&ldapserver.SearchRequest{
		BaseDN: "ou=groups,ou=test,dc=example,dc=com",
		Filter: "(cn=g1)",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 1, len(r.Entries))
	assertGroup(t, "g1", r.Entries[0])
}

func newTestUser(name string) User {
	return User{
		Name:   name,
		Groups: []string{"g1", "g2", "g3"},
		Attr: map[string][]string{
			"mail":        {fmt.Sprintf("%s@example.org", name)},
			"objectClass": {"some-class"},
		},
	}
}

func assertUser(t *testing.T, name string, entry *ldapserver.Entry) {
	assert.Equal(t, fmt.Sprintf("cn=%s,ou=people,ou=test,dc=example,dc=com", name), entry.DN)
	assert.Equal(t, 1, len(entry.GetAttributeValues("cn")))
	assert.Equal(t, name, entry.GetAttributeValue("cn"))
	assert.Equal(t, 1, len(entry.GetAttributeValues("mail")))
	assert.Equal(t, fmt.Sprintf("%s@example.org", name), entry.GetAttributeValue("mail"))
	assert.Equal(t, 2, len(entry.GetAttributeValues("objectClass")))
	assert.Contains(t, entry.GetAttributeValues("objectClass"), "inetOrgPerson")
	assert.Contains(t, entry.GetAttributeValues("objectClass"), "some-class")
	assert.Equal(t, 3, len(entry.GetAttributeValues("memberOf")))
	assert.Contains(t, entry.GetAttributeValues("memberOf"), "cn=g1,ou=groups,ou=test,dc=example,dc=com")
	assert.Contains(t, entry.GetAttributeValues("memberOf"), "cn=g2,ou=groups,ou=test,dc=example,dc=com")
	assert.Contains(t, entry.GetAttributeValues("memberOf"), "cn=g3,ou=groups,ou=test,dc=example,dc=com")
}

func newTestGroup(name string) Group {
	return Group{
		Name:    name,
		Members: []string{"u1", "u2", "u3"},
	}
}

func assertGroup(t *testing.T, name string, entry *ldapserver.Entry) {
	assert.Equal(t, fmt.Sprintf("cn=%s,ou=groups,ou=test,dc=example,dc=com", name), entry.DN)
	assert.Equal(t, 1, len(entry.GetAttributeValues("cn")))
	assert.Equal(t, name, entry.GetAttributeValue("cn"))
	assert.Equal(t, 1, len(entry.GetAttributeValues("objectClass")))
	assert.Contains(t, entry.GetAttributeValue("objectClass"), "groupOfNames")
	assert.Equal(t, 3, len(entry.GetAttributeValues("member")))
	assert.Contains(t, entry.GetAttributeValues("member"), "cn=u1,ou=people,ou=test,dc=example,dc=com")
	assert.Contains(t, entry.GetAttributeValues("member"), "cn=u2,ou=people,ou=test,dc=example,dc=com")
	assert.Contains(t, entry.GetAttributeValues("member"), "cn=u3,ou=people,ou=test,dc=example,dc=com")
}
