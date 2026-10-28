package adapter

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/domain"
)

// pageSize is how many items are requested at a time, and maxPages bounds how
// many pages are read for one listing so that an unexpectedly large account
// cannot turn one command into an unbounded number of requests.
const (
	pageSize = 100
	maxPages = 50
)

// Client performs operations against the StatusCake API.
type Client struct {
	logger *slog.Logger
	client *statuscake.Client
}

// New builds a Client from settings the adapter owns, so that nothing outside
// this package ever holds an SDK type. Replacing the SDK means replacing this
// package and nothing else.
func New(logger *slog.Logger, cfg Config) (*Client, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{logger: logger, client: client}, nil
}

// Window selects a slice of a resource's history. Before and After are
// instants; Limit is the most results wanted, where zero means the API's own
// default.
type Window struct {
	Limit  int
	Before gofp.Option[time.Time]
	After  gofp.Option[time.Time]
}

// collect reads every page of a listing.
func collect[T any](fetch func(page, limit int32) ([]T, statuscake.Pagination, error)) ([]T, error) {
	var all []T
	for page := int32(1); ; page++ {
		if page > maxPages {
			return nil, fmt.Errorf("%w: more than %d items", domain.ErrTruncated, maxPages*pageSize)
		}

		items, meta, err := fetch(page, pageSize)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)
		if page >= meta.PageCount {
			return all, nil
		}
	}
}

// translate turns an SDK error into the vocabulary of this package so that
// callers can react to a missing resource or a rejected token without importing
// the SDK. Anything else is passed on with context, and keeps its cause for
// [errors.As].
func translate(op string, err error) error {
	if err == nil {
		return nil
	}

	var apiErr statuscake.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusNotFound:
			return fmt.Errorf("%s: %w: %w", op, domain.ErrNotFound, err)
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("%s: %w: %w", op, domain.ErrUnauthorized, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

// apply calls set with the option's value when it has one.
func apply[T any](o gofp.Option[T], set func(T)) {
	if v, ok := o.TryUnwrap(); ok {
		set(v)
	}
}

// csv joins values for the API's comma separated filters.
func csv(values []string) string {
	return strings.Join(values, ",")
}

// instant converts an optional time to the Unix seconds the API expects.
func instant(o gofp.Option[time.Time], set func(int64)) {
	apply(o, func(t time.Time) { set(t.Unix()) })
}
