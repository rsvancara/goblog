package sessionmanager

import "fmt"

// Credentials Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Item User items
type Item struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

// User stores user information for a session
type User struct {
	Username     string `json:"username"`
	Items        []Item `json:"items"`
	IsAuth       bool   `json:"isauth"`
	IPAddress    string `json:"ipaddress"`
	City         string `json:"city"`
	TimeZone     string `json:"timezone"`
	Country      string `json:"country"`
	ASN          string `json:"asn"`
	Organization string `json:"organization"`
	SessionID    string `json:"sessionid"`
	TTL          int    `json:"ttl"`
}

// Set Item
func (u *User) SetItem(key string, value string) {
	var found bool

	for k := range u.Items {
		if u.Items[k].Key == key {
			u.Items[k].Val = value
			found = true
			break
		}
	}
	if !found {
		u.Items = append(u.Items, Item{key, value})
	}

	fmt.Println(u)
}

// Get Item
func (u *User) GetItem(key string) string {
	for k := range u.Items {
		if u.Items[k].Key == key {
			return u.Items[k].Val
		}
	}
	return ""
}

// Delete Item
func (u *User) DeleteItem(key string) {

	for i, v := range u.Items {

		if v.Key == key {

			u.Items = append(u.Items[:i], u.Items[i+1:]...)
		}
	}
}
