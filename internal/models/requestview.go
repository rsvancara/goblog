package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestView represents a request view
type RequestView struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	RequestViewID     string             `json:"requestviewid" bson:"requestviewid"`
	FunctionalBrowser string             `json:"functionalbrowser" bson:"functionalbrowser,omitempty"`
	SessionID         string             `json:"sessionid" bson:"sessionid"`
	OSVersion         string             `json:"osversion" bson:"osversion,omitempty"`
	OS                string             `json:"os" bson:"os,omitempty"`
	UserAgent         string             `json:"useragent" bson:"useragent,omitempty"`
	NavAppVersion     string             `json:"navappversion" bson:"navappversion,omitempty"`
	NavPlatform       string             `json:"navplatform" bson:"navplatform,omitempty"`
	NavBrowser        string             `json:"navbrowser" bson:"navbrowser,omitempty"`
	BrowserVersion    string             `json:"browserversion" bson:"browserversion,omitempty"`
	PTag              string             `json:"ptag" bson:"ptag"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	HeaderUserAgent   string             `json:"header_user_agent" bson:"header_user_agent"`
	IPAddress         string             `json:"ipaddress" bson:"ipaddress"`
	RequestURL        string             `json:"request_url" bson:"request_url"`
	City              string             `json:"city" bson:"city"`
	Country           string             `json:"country" bson:"country"`
	ASN               string             `json:"asn" bson:"asn"`
	Organization      string             `json:"orignization" bson:"organization"`
	RawRequest        string             `json:"raw_request" bson:"raw_request"`
}
