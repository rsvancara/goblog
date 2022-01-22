package models

import (
	"fmt"
	"time"
)

// Field The search terms you want to look for
type Field struct {
	Name      string `json:"name" bson:"name,omitempty"`           // Name name of the field to match on
	Operator  string `json:"operator" bson:"operator,omitempty"`   // Operator or, and, nand
	Value     string `json:"value" bson:"value,omitempty"`         // Value to match on
	Order     uint   `json:"order" bson:"order,omitempty"`         // Order of evaluation
	FieldType string `json:"fieldtype" bson:"fieldtype,omitempty"` // Field type, city, country, ip address are the current types
}

// Filter A type that represents a filter object
type Filter struct {
	FilterID          string    `json:"filter_id" bson:"filter_id,omitempty"`                   // Unique identifier
	Route             string    `json:"filter_type" bson:"route,omitempty"`                     // Filter route
	FilterAction      string    `json:"filter_action" bson:"filter_action,omitempty"`           // Allow Deny Redirect
	FilterRedirectURL string    `json:"filter_redirect_URL" bson:"filter_direct_URL,omitempty"` // Filter redirect URL REDIRECT is selected
	CreatedAt         time.Time `json:"created_at" bson:"created_at"`                           // CreatedAt date record was created
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at"`                           // UpdatedAt date record was updated
	Fields            []Field   `json:"fields" bson:"fields"`                                   // Matchers fields to match or not match on
	Name              string    `json:"name" bson:"name"`                                       // Name of the rule
	Description       string    `json:"description" bson:"description"`                         // Description of the rule
}

// GetFilterByID get a filter by id
func (f *Filter) GetFilterByID(id string) (Filter, error) {

	fmt.Println("Working")

	var filter Filter

	return filter, nil
}

// CreateFilter get a filter by id
func (f *Filter) CreateFilter() (Filter, error) {

	fmt.Println("Working")

	var filter Filter

	return filter, nil
}

// EditFilter update a filter
func (f *Filter) EditFilter() (Filter, error) {

	fmt.Println("Working")

	var filter Filter

	return filter, nil
}

// GetFiltersOrderByDate get a list of filters sorted by date
func GetFiltersOrderByDate() ([]Filter, error) {

	return nil, nil
}
