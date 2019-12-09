package mongo

import (
	"gopkg.in/mgo.v2"
)

// Session mongodb session
type Session struct {
	session *mgo.Session
}

// NewSession create new session
func NewSession(config *DbConfig) (*Session, error) {
	//var err error
	session, err := mgo.Dial(config.IP)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	return &Session{session}, err
}

// Copy copy session
func (s *Session) Copy() *mgo.Session {
	return s.session.Copy()
}

// Close close session
func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}
