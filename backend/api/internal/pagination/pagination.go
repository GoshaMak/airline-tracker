package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/google/uuid"
)

const (
	DefaultLimit = 50
	MaxLimit     = 100
)

var ErrInvalidParameters = errors.New("invalid pagination parameters")

type Params struct {
	Limit  int
	Cursor *uuid.UUID
}

type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

type cursorPayload struct {
	ID uuid.UUID `json:"id"`
}

func Parse(limitValue, cursorValue string) (Params, error) {
	params := Params{Limit: DefaultLimit}
	if limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit < 1 || limit > MaxLimit {
			return Params{}, ErrInvalidParameters
		}
		params.Limit = limit
	}

	if cursorValue == "" {
		return params, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(cursorValue)
	if err != nil {
		return Params{}, ErrInvalidParameters
	}
	var payload cursorPayload
	if err := json.Unmarshal(data, &payload); err != nil || payload.ID == uuid.Nil {
		return Params{}, ErrInvalidParameters
	}
	params.Cursor = &payload.ID
	return params, nil
}

func Paginate[T any](items []T, params Params, id func(T) uuid.UUID) Page[T] {
	ordered := append([]T(nil), items...)
	sort.Slice(ordered, func(i, j int) bool {
		return id(ordered[i]).String() < id(ordered[j]).String()
	})

	start := 0
	if params.Cursor != nil {
		cursor := params.Cursor.String()
		start = sort.Search(len(ordered), func(i int) bool {
			return id(ordered[i]).String() > cursor
		})
	}

	end := min(start+params.Limit, len(ordered))
	pageItems := ordered[start:end]
	hasMore := end < len(ordered)
	var nextCursor *string
	if hasMore && len(pageItems) > 0 {
		value := encode(id(pageItems[len(pageItems)-1]))
		nextCursor = &value
	}

	if pageItems == nil {
		pageItems = []T{}
	}
	return Page[T]{
		Items:      pageItems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

// FromFetched builds a page from an ordered database result fetched with
// FetchLimit. The result must contain only rows after Params.Cursor.
func FromFetched[T any](items []T, params Params, id func(T) uuid.UUID) Page[T] {
	hasMore := len(items) > params.Limit
	if hasMore {
		items = items[:params.Limit]
	}
	if items == nil {
		items = []T{}
	}
	var nextCursor *string
	if hasMore && len(items) > 0 {
		value := encode(id(items[len(items)-1]))
		nextCursor = &value
	}
	return Page[T]{Items: items, NextCursor: nextCursor, HasMore: hasMore}
}

func (p Params) FetchLimit() int {
	return p.Limit + 1
}

func (p Params) CursorValue() any {
	if p.Cursor == nil {
		return nil
	}
	return *p.Cursor
}

func Map[A, B any](page Page[A], convert func(A) B) Page[B] {
	items := make([]B, len(page.Items))
	for i, item := range page.Items {
		items[i] = convert(item)
	}
	return Page[B]{Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore}
}

func encode(id uuid.UUID) string {
	data, _ := json.Marshal(cursorPayload{ID: id})
	return base64.RawURLEncoding.EncodeToString(data)
}
