package models

import (
	"fmt"
	"time"
)

// Filter A type that represents a filter object
type Filter struct {
	FilterID           string    `json:"filter_id" bson:"filter_id,omitempty"`                         // Unique identifier
	FilterType         string    `json:"filter_type" bson:"filter_type,omitempty"`                     // Filter Type - PAGE or URL are the types
	FilterTypePageSlug string    `json:"filter_type_page_slug" bson:"filter_type_page_slug,omitempty"` // If the type is PAGE, choose page slug
	FilterTypeMatchURL string    `json:"filter_type_match_url" bson:"filter_type_match_url,omitempty"` // If type is MATCHURL then match on the URL regex
	FilterTypeURL      string    `json:"filter_type_url" bson:"filter_type_url,omitempty"`             // If type is MATCHURL then match on the URL regex
	FilterAction       string    `json:"filter_action" bson:"filter_action,omitempty"`                 // Allow Deny Redirect
	FilterRedirectURL  string    `json:"filter_redirect_URL" bson:"filter_direct_URL,omitempty"`       // Filter redirect URL REDIRECT is selected
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
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
