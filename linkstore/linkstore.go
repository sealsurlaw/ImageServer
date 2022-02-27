package linkstore

import "time"

type Link struct {
	FullFilename string
	ExpiresAt    *time.Time
}

type LinkStore interface {
	AddLink(token int64, link *Link) error
	GetLink(token int64) (*Link, error)
	DeleteLink(token int64) error
	Cleanup() error
}
