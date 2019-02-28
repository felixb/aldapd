package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
)

type BackendData struct {
	Users  []*User  `json:"users"`
	Groups []*Group `json:"groups"`
}

type localFileBackend struct {
	sync.RWMutex
	files        []string
	users        []User
	usersByName  map[string]*User
	usersByAttr  map[string][]User
	groups       []Group
	groupsByName map[string]*Group
	groupsByAttr map[string][]Group
}

func NewLocalFileBackend(files []string) (*localFileBackend, error) {
	b := &localFileBackend{files: files}
	return b, b.Reload()
}

func (b *localFileBackend) Check(username, password string) (bool, error) {
	b.RLock()
	defer b.RUnlock()

	if user, ok := b.usersByName[username]; !ok {
		return false, nil
	} else if user.Password == "" {
		return false, nil
	} else if len(user.Password) == 38 && strings.HasPrefix(user.Password, "{SSHA}") {
		return checkPasswordSSHA(username, password, user.Password)
	} else {
		log.Warningf("unknown password hash method for user %s", username)
		return false, nil
	}
}

func (b *localFileBackend) Users(filterKey, filterValue string) ([]User, error) {
	b.RLock()
	defer b.RUnlock()

	if filterKey == "" || filterValue == "" || filterValue == "*" {
		return b.users, nil
	} else if filterKey == "cn" {
		return []User{
			*b.usersByName[filterValue],
		}, nil
	} else {
		cacheKey := cacheKey(filterKey, filterValue)
		if users, ok := b.usersByAttr[cacheKey]; ok {
			log.Debugf("cache hit for filter %s on users", cacheKey)
			return users, nil
		} else {
			log.Debugf("cache miss for filter %s on users", cacheKey)
			users := b.filterUsers(filterKey, filterValue)
			b.usersByAttr[cacheKey] = users
			return users, nil
		}
	}
}

func (b *localFileBackend) filterUsers(filterKey, filterValue string) []User {
	if filterKey == "memberOf" {
		return b.filterUsersByGroup(filterValue)
	} else {
		return b.filterUsersByAttr(filterKey, filterValue)
	}
}

func (b *localFileBackend) filterUsersByGroup(name string) []User {
	if g, ok := b.groupsByName[name]; ok {
		users := make([]User, len(g.Members))
		for i, n := range g.Members {
			users[i] = *b.usersByName[n]
		}
		return users
	} else {
		return []User{}
	}
}

func (b *localFileBackend) filterUsersByAttr(attr, value string) []User {
	users := make([]User, 0)
	for _, u := range b.users {
		if values, ok := u.Attr[attr]; ok && contains(values, value) {
			users = append(users, u)
		}
	}
	return users
}

func (b *localFileBackend) Groups(filterKey, filterValue string) ([]Group, error) {
	b.RLock()
	defer b.RUnlock()

	if filterKey == "" || filterValue == "" || filterValue == "*" {
		return b.groups, nil
	} else if filterKey == "member" {
		cacheKey := cacheKey(filterKey, filterValue)
		if groups, ok := b.groupsByAttr[cacheKey]; ok {
			log.Debugf("cache hit for filter %s on groups", cacheKey)
			return groups, nil
		} else {
			log.Debugf("cache miss for filter %s on groups", cacheKey)
			groups := b.filterGroupsByMember(filterValue)
			b.groupsByAttr[cacheKey] = groups
			return groups, nil
		}
	} else {
		return []Group{}, nil
	}
}

func (b *localFileBackend) filterGroupsByMember(name string) []Group {
	groups := make([]Group, 0)
	for _, group := range b.groups {
		if contains(group.Members, name) {
			groups = append(groups, group)
		}
	}
	return groups
}

func (b *localFileBackend) Reload() error {
	var data BackendData
	usersByName := make(map[string]*User)
	groupsByName := make(map[string]*Group)
	for _, f := range b.files {
		log.Infof("loading users and groups data from %s", f)
		if content, err := ioutil.ReadFile(f); err != nil {
			return err
		} else if err := json.Unmarshal(content, &data); err != nil {
			return err
		} else {
			for _, user := range data.Users {
				log.Debugf("adding user %q", user.Name)
				usersByName[user.Name] = user
			}
			for _, group := range data.Groups {
				log.Debugf("adding group %q with %d members", group.Name, len(group.Members))
				groupsByName[group.Name] = group
			}
		}
	}

	for _, group := range groupsByName {
		for _, userName := range group.Members {
			if user, ok := usersByName[userName]; ok {
				log.Debugf("adding user %s to group %s", userName, group.Name)
				user.Groups = appendIfMissing(user.Groups, group.Name)
			}
		}
	}

	users := make([]User, len(usersByName))
	i := 0
	for _, v := range usersByName {
		users[i] = *v
		i++
	}
	groups := make([]Group, len(groupsByName))
	i = 0
	for _, v := range groupsByName {
		groups[i] = *v
		i++
	}

	b.Lock()
	b.users = users
	b.usersByName = usersByName
	b.usersByAttr = make(map[string][]User)
	b.groups = groups
	b.groupsByName = groupsByName
	b.groupsByAttr = make(map[string][]Group)
	b.Unlock()
	log.Infof("loaded %d users and %d groups", len(b.usersByName), len(b.groupsByName))
	return nil
}

func cacheKey(filterKey string, filterValue string) string {
	return fmt.Sprintf("(%s=%s)", filterKey, filterValue)
}

func checkPasswordSSHA(username, password, storedPassword string) (bool, error) {
	if byts, err := base64.StdEncoding.DecodeString(storedPassword[6:]); err != nil {
		log.Errorf("error decoding password for user %s", username)
		return false, err
	} else {
		salt := byts[20:]
		hash := byts[:20]
		check := sha1.Sum(append([]byte(password), salt...))
		return bytes.Equal(hash, check[:]), nil
	}
}
