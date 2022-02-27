package linkstore

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/errs"
)

type PostgresqlLinkStore struct {
	db    *sql.DB
	cache map[int64]*Link
}

func NewPostgresqlLinkStore(cfg *config.Config) (*PostgresqlLinkStore, error) {
	db, err := sql.Open("postgres", cfg.PostgresqlConfig.DatabaseString)
	if err != nil {
		return nil, err
	}
	if db.Ping() != nil {
		return nil, errs.ErrCannotConnectDatabase
	}

	cache := make(map[int64]*Link)
	linkStore := &PostgresqlLinkStore{db, cache}
	linkStore.buildTable()

	return linkStore, nil
}

func (s *PostgresqlLinkStore) buildTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			token BIGINT NOT NULL PRIMARY KEY UNIQUE,
			filepath TEXT NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresqlLinkStore) AddLink(token int64, link *Link) error {
	// check cache
	if s.cache[token] != nil {
		return errs.ErrTokenAlreadyExists
	}

	// check postgresql
	var count int
	err := s.db.QueryRow(`
		SELECT
			COUNT(*)
		FROM links
		WHERE
			token = $1
	`,
		token,
	).Scan(
		&count,
	)
	if err != nil {
		return err
	}
	if count != 0 {
		return errs.ErrTokenAlreadyExists
	}

	// perform insert asynchronously
	go func(cache *map[int64]*Link, token int64, link *Link) {
		// insert into cache
		s.cache[token] = link

		// insert into postgresql
		_, err = s.db.Exec(`
			INSERT INTO links (
				token,
				filepath,
				expires_at
			) VALUES ($1, $2, $3)
		`,
			token,
			link.FullFilename,
			link.ExpiresAt,
		)
	}(&s.cache, token, link)

	return nil
}

func (s *PostgresqlLinkStore) GetLink(token int64) (*Link, error) {
	link := &Link{}
	if s.cache[token] != nil {
		// check cache
		link = s.cache[token]
	} else {
		// check postgresql
		err := s.db.QueryRow(`
		SELECT
			filepath,
			expires_at
		FROM links WHERE
			token = $1
	`,
			token,
		).Scan(
			&link.FullFilename,
			&link.ExpiresAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errs.ErrTokenNotFound
			}
			return nil, err
		}

		// populate cache
		s.cache[token] = link
	}

	// remove if expired
	if time.Now().After(*link.ExpiresAt) {
		s.DeleteLink(token)
		return nil, errs.ErrTokenExpired
	}

	return link, nil
}

func (s *PostgresqlLinkStore) DeleteLink(token int64) error {
	// remove from cache
	delete(s.cache, token)

	// remove from postgresql
	res, err := s.db.Exec(`
		DELETE FROM links WHERE
			token = $1
	`,
		token,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.ErrTokenNotFound
	}

	return nil
}

func (s *PostgresqlLinkStore) Cleanup() error {
	// cleaup cache
	for token, link := range s.cache {
		if time.Now().After(*link.ExpiresAt) {
			delete(s.cache, token)
		}
	}

	// cleanup postgresql
	_, err := s.db.Exec(`
		DELETE FROM links WHERE
			expires_at < $1
	`,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}
