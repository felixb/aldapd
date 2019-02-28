package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	validTestConfig = `{
	"users": [
		{"name":"u1", "attr":{"a1":["v1"]}, "password":"some-password"},
		{"name":"u2"}
],
	"groups": [
		{"name":"g1", "member": ["u1","u2"]},
		{"name":"g2", "member": ["u1"]}
]
}`
)

func TestNewLocalFileBackend_missing_file(t *testing.T) {
	_, err := NewLocalFileBackend([]string{"/tmp/missing"})
	assert.Error(t, err)
}

func TestNewLocalFileBackend_not_json(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-not-json")
	defer os.Remove(f.Name())
	f.WriteString("not json")

	_, err := NewLocalFileBackend([]string{f.Name()})
	assert.Error(t, err)
}

func TestNewLocalFileBackend(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	assert.Equal(t, 1, len(b.files))
	assert.Equal(t, f.Name(), b.files[0])

	assert.Equal(t, 2, len(b.users))
	assert.Equal(t, 2, len(b.usersByName))

	assert.Equal(t, "u1", b.usersByName["u1"].Name)
	assert.ElementsMatch(t, []string{"g1", "g2"}, b.usersByName["u1"].Groups)
	assert.Equal(t, 1, len(b.usersByName["u1"].Attr))
	assert.Equal(t, "some-password", b.usersByName["u1"].Password)

	assert.Equal(t, "u2", b.usersByName["u2"].Name)
	assert.ElementsMatch(t, []string{"g1"}, b.usersByName["u2"].Groups)

	assert.Equal(t, 2, len(b.groups))
	assert.Equal(t, 2, len(b.groupsByName))

	assert.Equal(t, "g1", b.groupsByName["g1"].Name)
	assert.ElementsMatch(t, []string{"u1", "u2"}, b.groupsByName["g1"].Members)

	assert.Equal(t, "g2", b.groupsByName["g2"].Name)
	assert.ElementsMatch(t, []string{"u1"}, b.groupsByName["g2"].Members)
}

func TestLocalFileBackend_Reload(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	lu := len(b.usersByName)
	lg := len(b.groupsByName)
	assert.True(t, lu > 0)
	assert.True(t, lg > 0)

	f.WriteString("not-json")
	assert.Error(t, b.Reload())
	assert.Equal(t, lu, len(b.usersByName))
	assert.Equal(t, lg, len(b.groupsByName))

	os.Remove(f.Name())
	assert.Error(t, b.Reload())
	assert.Equal(t, lu, len(b.usersByName))
	assert.Equal(t, lg, len(b.groupsByName))

	f, _ = os.Create(f.Name())
	f.WriteString(`{}`)
	assert.NoError(t, b.Reload())
	assert.Equal(t, 0, len(b.usersByName))
	assert.Equal(t, 0, len(b.groupsByName))
}

func TestLocalFileBackend_Users(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	users, err := b.Users("", "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.ElementsMatch(t, []string{"u1", "u2"}, nameOfUsers(users))
}

func TestLocalFileBackend_Users_filterByCn(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	users, err := b.Users("cn", "u1")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "u1", users[0].Name)
	assert.ElementsMatch(t, []string{"g1", "g2"}, users[0].Groups)
	assert.Equal(t, []string{"v1"}, users[0].Attr["a1"])
}

func TestLocalFileBackend_Users_filterByGroupMultiple(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	for range []int{1, 2, 3} {
		users, err := b.Users("memberOf", "g1")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.ElementsMatch(t, []string{"u1", "u2"}, nameOfUsers(users))
	}
}

func TestLocalFileBackend_Users_filterByGroupSingle(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	for range []int{1, 2, 3} {
		users, err := b.Users("memberOf", "g2")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, "u1", users[0].Name)
	}
}

func TestLocalFileBackend_Users_filterByGroupZero(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	for range []int{1, 2, 3} {
		users, err := b.Users("memberOf", "g3")
		assert.NoError(t, err)
		assert.Equal(t, 0, len(users))
	}
}

func TestLocalFileBackend_Users_filterByAttr(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	for range []int{1, 2, 3} {
		users, err := b.Users("a1", "v1")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, "u1", users[0].Name)
	}
}

func TestLocalFileBackend_Groups(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	groups, err := b.Groups("", "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))
	assert.ElementsMatch(t, []string{"g1", "g2"}, nameOfGroups(groups))
}

func TestLocalFileBackend_Groups_filterByMemberMultiple(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	groups, err := b.Groups("member", "u1")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))
	assert.ElementsMatch(t, []string{"g1", "g2"}, nameOfGroups(groups))
}

func TestLocalFileBackend_Groups_filterByMemberSingle(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	groups, err := b.Groups("member", "u2")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, "g1", groups[0].Name)
}

func TestLocalFileBackend_Groups_filterByMemberZero(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	groups, err := b.Groups("member", "u3")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(groups))
}

func TestLocalFileBackend_Groups_filterByAttr(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "aldapd-config")
	defer os.Remove(f.Name())
	f.WriteString(validTestConfig)

	b, err := NewLocalFileBackend([]string{f.Name()})
	assert.NoError(t, err)

	groups, err := b.Groups("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(groups))
}

func TestLocalFileBackend_Check_invalids(t *testing.T) {
	b := &localFileBackend{
		usersByName: map[string]*User{"u1": {}},
	}

	cases := [][]string{
		{"u1", ""},
		{"u1", "foo"},
		{"u2", ""},
		{"u2", "bar"},
	}

	for _, e := range cases {
		r, err := b.Check(e[0], e[1])
		assert.NoError(t, err)
		assert.False(t, r, "for '%s' / '%s'", e[0], e[1])
	}
}

func TestLocalFileBackend_Check(t *testing.T) {
	cases := map[string]string{
		"foo": "{SSHA}hNsogC9IKy6CFkQzyDSMPmOlAnxcc27o",
		"öäü": "{SSHA}9SP8txPWXqn1D7osBhKl6lCGHYTthMJe",
	}

	for k, v := range cases {
		b := &localFileBackend{
			usersByName: map[string]*User{"u1": {Password: v}},
		}

		r, err := b.Check("u1", k)
		assert.NoError(t, err)
		assert.True(t, r, "for '%s'", k)

		r, err = b.Check("u1", "something-other")
		assert.NoError(t, err)
		assert.False(t, r, "for '%s'", k)
	}
}

func nameOfUsers(users []User) []string {
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}
	return names
}

func nameOfGroups(groups []Group) []string {
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return names
}
