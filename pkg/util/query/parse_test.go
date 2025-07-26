package query

import (
	"reflect"
	"testing"
)

func TestParseQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *QueryOptions
		err      bool
	}{
		{
			name:  "empty query",
			query: "",
			expected: &QueryOptions{
				Select:     nil,
				HasSelect:  false,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
			},
		},
		{
			name:  "simple select and alias",
			query: "SELECT foo, bar as baz",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "foo", Reducer: ""},
					{Field: ".bar", Alias: "baz", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
			},
		},
		{
			name:  "where and limit",
			query: "where count>0 limit 5",
			expected: &QueryOptions{
				Select:     nil,
				HasSelect:  false,
				Where:      "count>0",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      5,
				HasLimit:   true,
			},
		},
		{
			name:  "order by asc and desc",
			query: "order by foo desc, bar ASC",
			expected: &QueryOptions{
				Select:    nil,
				HasSelect: false,
				Where:     "",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".foo", Alias: "foo", Reducer: ""},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".bar", Alias: "bar", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      -1,
				HasLimit:   false,
			},
		},
		{
			name:  "combined full query",
			query: "SELECT sum(x) as total, y WHERE y>1 ORDER BY x DESC, y LIMIT 10",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".x", Alias: "total", Reducer: "sum"},
					{Field: ".y", Alias: "y", Reducer: ""},
				},
				HasSelect: true,
				Where:     "y>1",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".x", Alias: "total", Reducer: "sum"},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".y", Alias: "y", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      10,
				HasLimit:   true,
			},
		},
		{
			name:  "order by alias",
			query: "SELECT foo as f, bar as b ORDER BY f DESC, b",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "f", Reducer: ""},
					{Field: ".bar", Alias: "b", Reducer: ""},
				},
				HasSelect: true,
				Where:     "",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".foo", Alias: "f", Reducer: ""},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".bar", Alias: "b", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      -1,
				HasLimit:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueryString(tt.query)
			if (err != nil) != tt.err {
				t.Fatalf("unexpected error status: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseQueryString(%q) =\n  %#v\nexpected\n  %#v", tt.query, got, tt.expected)
			}
		})
	}
}

func TestParseSelectClauseFunctionOptionalParentheses(t *testing.T) {
	tests := []struct {
		input string
		want  SelectOption
	}{
		{"len hello", SelectOption{Field: ".hello", Reducer: "len", Alias: "hello"}},
		{"len(hello)", SelectOption{Field: ".hello", Reducer: "len", Alias: "hello"}},
		{"sum value as total", SelectOption{Field: ".value", Reducer: "sum", Alias: "total"}},
		{"sum(value) as total", SelectOption{Field: ".value", Reducer: "sum", Alias: "total"}},
	}

	for _, tc := range tests {
		got := parseSelectClause(tc.input)
		if len(got) != 1 {
			t.Errorf("parseSelectClause(%q) returned %d opts, want 1", tc.input, len(got))
			continue
		}
		if !reflect.DeepEqual(got[0], tc.want) {
			t.Errorf("parseSelectClause(%q)[0] = %+v, want %+v", tc.input, got[0], tc.want)
		}
	}
}
