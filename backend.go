package main

type Backender interface {
	Check(username, password string) (bool, error)
	Users(filterKey, filterValue string) ([]User, error)
	Groups(filterKey, filterValue string) ([]Group, error)
	Reload() error
}

type User struct {
	Name     string              `json:"name"`
	Groups   []string            `json:",-"`
	Attr     map[string][]string `json:"attr"`
	Password string              `json:"password"`
}

type Group struct {
	Name    string   `json:"name"`
	Members []string `json:"member"`
}
