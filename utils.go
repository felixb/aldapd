package main

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/op/go-logging"
)

func contains(arr []string, s string) bool {
	for _, e := range arr {
		if e == s {
			return true
		}
	}
	return false
}

func appendIfMissing(arr []string, s string) []string {
	for _, e := range arr {
		if e == s {
			return arr
		}
	}
	r := append(arr, s)
	sort.Strings(r)
	return r
}

func cn2dn(baseDn, cn string) string {
	return fmt.Sprintf("cn=%s,%s", cn, baseDn)
}

func cns2dns(baseDn string, cns []string) []string {
	dns := make([]string, len(cns))
	for i, cn := range cns {
		dns[i] = cn2dn(baseDn, cn)
	}
	return dns
}

func dn2cn(baseDn, dn string) (string, bool) {
	re := regexp.MustCompile(fmt.Sprintf("^cn=([^,]+),.*%s$", baseDn))
	found := re.FindStringSubmatch(dn)
	if len(found) < 2 {
		log.Warningf("failed to convert dn=%s to username, baseDn=%s", dn, baseDn)
		return "", false
	}
	return found[1], true
}

func redactNonEmpty(s string) string {
	if s != "" {
		return logging.Redact(s)
	} else {
		return ""
	}
}
