package router

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ktt-ol/mqlux/internal/mqlux"
)

func TestSortByPath(t *testing.T) {
	for _, test := range []struct {
		input []handler
		want  []handler
	}{
		{
			input: []handler{},
			want:  []handler{},
		},
		{
			input: []handler{{path: []string{"a", "b", "c"}}},
			want:  []handler{{path: []string{"a", "b", "c"}}},
		},
		{
			input: []handler{{path: []string{"c"}}, {path: []string{"b"}}, {path: []string{"a"}}},
			want:  []handler{{path: []string{"a"}}, {path: []string{"b"}}, {path: []string{"c"}}},
		},
		{
			input: []handler{
				{path: []string{"a", "b", "c"}},
				{path: []string{"a"}},
				{path: []string{"a", "#"}},
				{path: []string{"a", "a"}},
				{path: []string{"c"}},
				{path: []string{"a", "b", "c", "d"}},
				{path: []string{"c"}},
				{path: []string{"#"}},
				{path: []string{"a", "b", "d"}},
				{path: []string{"b"}},
			},
			want: []handler{
				{path: []string{"#"}},
				{path: []string{"a"}},
				{path: []string{"a", "#"}},
				{path: []string{"a", "a"}},
				{path: []string{"a", "b", "c"}},
				{path: []string{"a", "b", "c", "d"}},
				{path: []string{"a", "b", "d"}},
				{path: []string{"b"}},
				{path: []string{"c"}},
				{path: []string{"c"}},
			},
		},
	} {
		sort.Sort(byPath(test.input))
		if !reflect.DeepEqual(test.input, test.want) {
			t.Errorf("unexpected sort order %v != %v", test.input, test.want)
		}
	}
}

type dummyHandler struct {
	path string
}

func (h *dummyHandler) Receive(msg mqlux.Message) {
}

func (h *dummyHandler) Topic() string {
	return h.path
}

func TestRouter(t *testing.T) {
	for _, test := range []struct {
		input  []string
		search string
		want   []int
	}{
		{
			input: []string{
				"a/b/c",
				"a",
				"a/b",
				"a/b",
				"a/b/c/d",
				"c",
				"a/b/#",
				"a/b/d",
				"b",
			},
			search: "a/b",
			want: []int{
				2, // "a/b"
				3, // "a/b"
				6, // "a/b/#"
			},
		},
		{
			input: []string{
				"c/d/e",
				"c/d/e",
				"b",
			},
			search: "c/d/e",
			want: []int{
				0, // "c/d/e"
				1, // "c/d/e"
			},
		},
		{
			input: []string{
				"a/b/c/d",
				"c",
				"c/d/e",
				"b",
			},
			search: "c",
			want: []int{
				1, //"c"
			},
		},
		{
			input:  []string{"b"},
			search: "a",
			want:   nil,
		},
		{
			input:  []string{"b"},
			search: "c",
			want:   nil,
		},
		{
			input: []string{
				"a/b/c/d",
				"c/#",
				"c/d/e",
				"b",
				"#",
			},
			search: "c",
			want: []int{
				1, // "c/#"
				4, // "#"
			},
		},
		{
			input: []string{
				"/a/#",
				"/a/a/x",
				"/a/x",
				"/b/a/x",
				"/b/x",
			},
			search: "/a/a/x",
			want: []int{
				1, // "/a/a/x"
				0, // "/a/#"
			},
		},
	} {
		r := New()
		hs := []Receiver{}
		for _, path := range test.input {
			h := dummyHandler{path: path}
			r.Add(path, &h)
			hs = append(hs, &h)
		}
		result := r.Find(test.search)
		if len(result) != len(test.want) {
			t.Errorf("unexpected result %v != %v", result, test.want)
			continue
		}
		for i := range result {
			if !reflect.DeepEqual(result[i], hs[test.want[i]]) {
				t.Errorf("unexpected result[%d] %v != %v", i, result, test.want)
			}
		}
	}
}

func BenchmarkRouter(b *testing.B) {
	r := New()
	randomPath := func() string {
		length := rand.Intn(4) + 2
		path := make([]string, length)
		for i := range path {
			path[i] = string('a'+rand.Intn(26)) + string('a'+rand.Intn(26))
		}
		if rand.Intn(10) == 0 {
			path[length-1] = "#"
		}
		return strings.Join(path, "/")
	}
	for i := 0; i < 100000; i++ {
		r.Add(randomPath(), nil)
	}
	// first Find is slow as it sorts our topics
	r.Find(randomPath())

	b.ResetTimer()
	b.ReportAllocs()

	found := 0
	topics := 0
	for n := 0; n < b.N; n++ {
		search := randomPath()
		results := len(r.Find(search))
		if results > 0 {
			found++
			topics += results
		}
	}
	b.Log(b.N, found, float64(topics)/float64(found))
}

func TestHasPrefix(t *testing.T) {
	for _, test := range []struct {
		path     []string
		prefix   []string
		isPrefix bool
	}{
		{path: nil, prefix: []string{"a"}, isPrefix: false},
		{path: []string{"a"}, prefix: []string{"a"}, isPrefix: true},
		{path: []string{"a", "b"}, prefix: []string{"a"}, isPrefix: true},
		{path: []string{"b"}, prefix: []string{"a"}, isPrefix: false},
		{path: []string{"a", "b", "c"}, prefix: []string{"a", "c"}, isPrefix: false},
		{path: []string{"a", "b"}, prefix: []string{"a", "b", "c"}, isPrefix: false},
	} {
		if hasPrefix(test.path, test.prefix) != test.isPrefix {
			t.Errorf("hasPrefix for %v %v != %v",
				test.path, test.prefix, test.isPrefix)
		}
	}
}
