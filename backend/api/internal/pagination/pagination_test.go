package pagination

import (
	"testing"

	"github.com/google/uuid"
)

type testItem struct {
	ID uuid.UUID
}

func TestPaginateRoundTrip(t *testing.T) {
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	}
	items := []testItem{{ID: ids[0]}, {ID: ids[1]}, {ID: ids[2]}}

	first := Paginate(items, Params{Limit: 2}, func(item testItem) uuid.UUID { return item.ID })
	if len(first.Items) != 2 || !first.HasMore || first.NextCursor == nil {
		t.Fatalf("first page = %+v", first)
	}
	if first.Items[0].ID != ids[1] || first.Items[1].ID != ids[2] {
		t.Fatalf("items are not ordered: %+v", first.Items)
	}

	params, err := Parse("2", *first.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	second := Paginate(items, params, func(item testItem) uuid.UUID { return item.ID })
	if len(second.Items) != 1 || second.Items[0].ID != ids[0] || second.HasMore || second.NextCursor != nil {
		t.Fatalf("second page = %+v", second)
	}
}

func TestParseRejectsInvalidParameters(t *testing.T) {
	for _, test := range []struct {
		limit  string
		cursor string
	}{
		{limit: "0"},
		{limit: "101"},
		{limit: "not-a-number"},
		{cursor: "not-a-cursor"},
	} {
		if _, err := Parse(test.limit, test.cursor); err == nil {
			t.Fatalf("Parse(%q, %q) succeeded", test.limit, test.cursor)
		}
	}
}

func TestFromFetchedUsesLookaheadRow(t *testing.T) {
	items := []testItem{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002")},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003")},
	}
	page := FromFetched(items, Params{Limit: 2}, func(item testItem) uuid.UUID { return item.ID })
	if len(page.Items) != 2 || !page.HasMore || page.NextCursor == nil {
		t.Fatalf("page = %+v", page)
	}
}
