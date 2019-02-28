# atto ldap daemon

This is nearly the smallest ldap daemon you could run.
It comes with a static backend configurable in json format.
There is only support for read access implemented in `aldapd`.

## User case

`aldapd` was built as part of a distributed SSO solution.
It is meant to be deployed on every machine providing login locally.
There ought to be a central backend for your user credentials allowing changing passwords and managing groups.
On any change a snapshot of the database is to be distributed to all `aldapd` instances. 

## Limits

`aldapd` supports only a very small set of LDAP requests:

* bind
  Binding to `aldapd` is optionally allowd anonymously with empty bindDN and password.
  It also supports binding as a user with salted and sha1 hashed passwords following SSHA standards.
* search
  `aldapd` responses to queries for users under `ou=people,${baseDN}` and groups under `ou=groups,${baseDN}`.
  It supports only filters with a single expression like `(objectClass=*)` or `(cn=mandy)`.
  Other queries will result in empty results. 

## Configuration of your application

Configure your application with similar settings:

* baseDn `dc=felixb,dc=github,dc=com`
* userSearchBase `ou=people`
* userSearch `cn={0}`
* userMembershipAttributeName `memberOf`
* groupSearchBase `ou=groups`
* groupSearchFilter `cn={0}`
* groupMembershipFilter `member={0}`


## Reloading the config

`aldapd` reads the backend config once on startup and keeps a copy in memory.
To reload the config, simply send a `SIGUSR1` to the process.
`aldapd` reads the config again, checks it and replaces the old in memory copy with the new one.

## Example config

The following example configuration shows two users and two groups:

```json
{
	"users": [
		{"name":"jacqueline", "attr":{"mail":["jacqueline@example.org"]}, "password":"{SSHA}hNsogC9IKy6CFkQzyDSMPmOlAnxcc27o"},
		{"name":"kevin", "attr":{"mail":["kevin@example.org"]}, "password": "{SSHA}9SP8txPWXqn1D7osBhKl6lCGHYTthMJe"}
    ],
	"groups": [
		{"name":"developer", "member": ["jacqueline","kevin"]},
		{"name":"admin", "member": ["jacqueline"]}
    ]
}
```

This results in the following LDAP outputs:

```bash
$ ldapsearch -b ou=people,dc=felixb,dc=github,dc=com

# jacqueline, people, felixb.github.com
dn: cn=jacqueline,ou=people,dc=felixb,dc=github,dc=com
mail: jacqueline@example.org
cn: jacqueline
objectClass: inetOrgPerson
memberOf: cn=admin,ou=groups,dc=felixb,dc=github,dc=com
memberOf: cn=developer,ou=groups,dc=felixb,dc=github,dc=com

# kevin, people, felixb.github.com
dn: cn=kevin,ou=people,dc=felixb,dc=github,dc=com
mail: kevin@example.org
cn: kevin
objectClass: inetOrgPerson
memberOf: cn=developer,ou=groups,dc=felixb,dc=github,dc=com
```

```bash
$ ldapsearch -b ou=groups,dc=felixb,dc=github,dc=com  

# developer, groups, felixb.github.com
dn: cn=developer,ou=groups,dc=felixb,dc=github,dc=com
cn: developer
member: cn=jacqueline,ou=people,dc=felixb,dc=github,dc=com
member: cn=kevin,ou=people,dc=felixb,dc=github,dc=com
objectClass: groupOfNames

# admin, groups, felixb.github.com
dn: cn=admin,ou=groups,dc=felixb,dc=github,dc=com
cn: admin
member: cn=jacqueline,ou=people,dc=felixb,dc=github,dc=com
objectClass: groupOfNames
```

## Extending LDAP objects

`aldapd` allows extending the LDAP objects representing users by adding additional classes or attributes:

```json
{
  "name":"jason",
  "attr":{
    "objectClass": ["posixAccount"],
    "uid": ["jason"],
    "uidNumber": ["1234"],
    "gidNumber": ["1234"],
    "homeDirectory": ["/home/jason"]
  }
}
```

This results in

```bash
# jason, people, felixb.github.com
dn: cn=jason,ou=people,dc=felixb,dc=github,dc=com
uidNumber: 1234
gidNumber: 1234
homeDirectory: /home/jason
uid: jason
cn: jason
objectClass: inetOrgPerson
objectClass: posixAccount
```

## Extending `aldapd`

`aldapd` is designed to allow replacing the backend easily.
An `aldapd` backend just needs to implement the `Backender` interface.
Then swap the LocalFileBackend out and put your own implementation in.

There are a bunch of backend I can think of right away:

* backend config written some other format like yaml or moml
* backend config stored in AWS S3
* backend config stored in AWS dynamodb
