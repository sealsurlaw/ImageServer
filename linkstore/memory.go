package linkstore

import (
	"time"

	"github.com/sealsurlaw/ImageServer/errs"
)

type MemoryLinkStore struct {
	links map[int64]*Link
}

func NewMemoryLinkStore() *MemoryLinkStore {
	links := make(map[int64]*Link)
	return &MemoryLinkStore{
		links: links,
	}
}

func (s *MemoryLinkStore) AddLink(token int64, link *Link) error {
	if s.links[token] != nil {
		return errs.ErrTokenAlreadyExists
	}

	s.links[token] = link

	return nil
}

func (s *MemoryLinkStore) DeleteLink(token int64) error {
	if s.links[token] == nil {
		return errs.ErrTokenNotFound
	}

	delete(s.links, token)

	return nil
}

func (s *MemoryLinkStore) GetLink(token int64) (*Link, error) {
	if s.links[token] == nil {
		return nil, errs.ErrTokenNotFound
	}

	if time.Now().After(*s.links[token].ExpiresAt) {
		s.DeleteLink(token)
		return nil, errs.ErrTokenExpired
	}

	return s.links[token], nil
}
