package main

import (
	"net"
	"strings"

	"github.com/mark-rushakoff/ldapserver"
)

func (s *Server) bind(bindDn, bindSimplePw string, conn net.Conn) (ldapserver.LDAPResultCode, error) {
	log.Debugf("bind request: bindDn=%s, bindSimplePw=%s", bindDn, redactNonEmpty(bindSimplePw))
	if bindDn == "" && bindSimplePw == "" {
		if s.config.allowAnonBind {
			return ldapserver.LDAPResultSuccess, nil
		} else {
			return ldapserver.LDAPResultInvalidCredentials, nil
		}
	} else if username, ok := dn2cn(s.config.baseDn, strings.ToLower(bindDn)); !ok {
		return ldapserver.LDAPResultInvalidCredentials, nil
	} else if ok, err := s.backend.Check(username, bindSimplePw); !ok {
		return ldapserver.LDAPResultInvalidCredentials, err
	} else {
		return ldapserver.LDAPResultSuccess, err
	}
}
