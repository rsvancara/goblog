package models

// RequestView represents a request view
type RequestView struct {
	FunctionalBrowser string `json:"functionalbrowser"`
	Sessionid         string `json:"sessionid"`
	OSVersion         string `json:"osversion"`
	OS                string `json:"os"`
	UserAgent         string `json:"useragent"`
	NavAppVersion     string `json:"navappversion"`
	NavPlatform       string `json:"navplatform"`
	NavBrowser        string `json:"navbrowser"`
	BrowserVersion    string `json:"browserversion"`
	PTag              string `json:"ptag"`
}

//CreateRequestView create a new requestview
func (r *RequestView) CreateRequestView() error {
	return nil
}

// GetRequestViewByPageID get a requestview by pageid
func (r *RequestView) GetRequestViewByPageID() error {
	return nil
}

// GetRequestViewsBySessionID get a list of requestviews by sessionid
func GetRequestViewsBySessionID(id string) ([]RequestView, error) {
	return nil, nil
}
