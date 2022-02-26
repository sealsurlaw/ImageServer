package linkstore

import "time"

type Link struct {
	FullFilename string
	ExpiresAt    *time.Time
}

type LinkStore interface {
	AddLink(token int64, link *Link) error
	DeleteLink(token int64) error
	GetLink(token int64) (*Link, error)
}
