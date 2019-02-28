package main

import (
	"fmt"
	"net"
	"regexp"

	"github.com/mark-rushakoff/ldapserver"
)

func (s *Server) search(boundDn string, req ldapserver.SearchRequest, conn net.Conn) (ldapserver.ServerSearchResult, error) {
	log.Debugf("search request: bindDn=%s, baseDn=%s, filter=%s", boundDn, req.BaseDN, req.Filter)

	switch req.BaseDN {
	case s.config.peopleDn:
		return s.searchUsers(req)
	case s.config.groupsDn:
		return s.searchGroups(req)
	default:
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultInsufficientAccessRights,
		}, nil
	}
}

func (s *Server) searchUsers(req ldapserver.SearchRequest) (ldapserver.ServerSearchResult, error) {
	log.Debug("search people request")

	if filterKey, filterValue, err := parseFilter(req.Filter); err != nil {
		log.Errorf("error parsing search filter: %s", err.Error())
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultOperationsError,
		}, err
	} else if users, err := s.backend.Users(filterKey, filterValue); err != nil {
		log.Errorf("error getting users from backend: %s", err.Error())
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultOperationsError,
		}, err
	} else {
		return ldapserver.ServerSearchResult{
			Entries:    users2entries(users, s.config.peopleDn, s.config.groupsDn),
			ResultCode: ldapserver.LDAPResultSuccess,
		}, nil
	}
}

func (s *Server) searchGroups(req ldapserver.SearchRequest) (ldapserver.ServerSearchResult, error) {
	log.Debug("search groups request")

	if filterKey, filterValue, err := parseFilter(req.Filter); err != nil {
		log.Errorf("error parsing search filter: %s", err.Error())
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultOperationsError,
		}, err
	} else if groups, err := s.backend.Groups(filterKey, filterValue); err != nil {
		log.Errorf("error getting groups from backend: %s", err.Error())
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultOperationsError,
		}, err
	} else if groups == nil {
		log.Warning("backend did not return any groups")
		return ldapserver.ServerSearchResult{
			ResultCode: ldapserver.LDAPResultSuccess,
		}, nil

	} else {
		return ldapserver.ServerSearchResult{
			Entries:    groups2entries(groups, s.config.peopleDn, s.config.groupsDn),
			ResultCode: ldapserver.LDAPResultSuccess,
		}, nil
	}
}

func parseFilter(filter string) (string, string, error) {
	re := regexp.MustCompile(`^\((\w+)=([^)]+)\)$`)
	m := re.FindStringSubmatch(filter)
	if len(m) < 3 {
		return "", "", fmt.Errorf("unsupported search filter '%s', only filter in form '(key=value)' allowed", filter)
	} else {
		filterKey := m[1]
		filterValue := m[2]
		if filterKey == "objectClass" && filterValue == "*" {
			return "", "", nil
		} else {
			return filterKey, filterValue, nil
		}
	}
}

func appendAttr(attr []*ldapserver.EntryAttribute, name string, values ...string) []*ldapserver.EntryAttribute {
	if len(values) > 0 {
		return append(attr, &ldapserver.EntryAttribute{Name: name, Values: values})
	} else {
		return attr
	}
}

func user2entry(user *User, peopleDn, groupDn string) *ldapserver.Entry {
	attr := make([]*ldapserver.EntryAttribute, 0)
	classes := make([]string, 0)
	for k, v := range user.Attr {
		if k == "objectClass" {
			classes = append(classes, v...)
		} else {
			attr = appendAttr(attr, k, v...)
		}
	}
	attr = appendAttr(attr, "cn", user.Name)
	attr = appendAttr(attr, "objectClass", appendIfMissing(classes, "inetOrgPerson")...)
	attr = appendAttr(attr, "memberOf", cns2dns(groupDn, user.Groups)...)

	return &ldapserver.Entry{
		DN:         cn2dn(peopleDn, user.Name),
		Attributes: attr,
	}
}
func users2entries(users []User, peopleDn, groupDn string) []*ldapserver.Entry {
	entries := make([]*ldapserver.Entry, len(users))
	for i, user := range users {
		entries[i] = user2entry(&user, peopleDn, groupDn)
	}
	return entries
}

func group2entry(group *Group, peopleDn, groupDn string) *ldapserver.Entry {
	attr := make([]*ldapserver.EntryAttribute, 0)
	attr = appendAttr(attr, "cn", group.Name)
	attr = appendAttr(attr, "member", cns2dns(peopleDn, group.Members)...)
	attr = appendAttr(attr, "objectClass", "groupOfNames")

	return &ldapserver.Entry{
		DN:         cn2dn(groupDn, group.Name),
		Attributes: attr,
	}
}

func groups2entries(groups []Group, peopleDn, groupDn string) []*ldapserver.Entry {
	entries := make([]*ldapserver.Entry, len(groups))
	for i, group := range groups {
		entries[i] = group2entry(&group, peopleDn, groupDn)
	}
	return entries
}
