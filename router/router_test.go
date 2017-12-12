package router

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestSortByPath(t *testing.T) {
	for _, test := range []struct {
		input []topic
		want  []topic
	}{
		{
			input: []topic{},
			want:  []topic{},
		},
		{
			input: []topic{{path: []string{"a", "b", "c"}}},
			want:  []topic{{path: []string{"a", "b", "c"}}},
		},
		{
			input: []topic{{path: []string{"c"}}, {path: []string{"b"}}, {path: []string{"a"}}},
			want:  []topic{{path: []string{"a"}}, {path: []string{"b"}}, {path: []string{"c"}}},
		},
		{
			input: []topic{
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
			want: []topic{
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

func TestRouter(t *testing.T) {
	for _, test := range []struct {
		input  [][]string
		search []string
		want   [][]string
	}{
		{
			input: [][]string{
				[]string{"a", "b", "c"},
				[]string{"a"},
				[]string{"a", "b"},
				[]string{"a", "b"},
				[]string{"a", "b", "c", "d"},
				[]string{"c"},
				[]string{"a", "b", "#"},
				[]string{"a", "b", "d"},
				[]string{"b"},
			},
			search: []string{"a", "b"},
			want: [][]string{
				[]string{"a", "b"},
				[]string{"a", "b"},
				[]string{"a", "b", "#"},
			},
		},
		{
			input: [][]string{
				[]string{"c", "d", "e"},
				[]string{"c", "d", "e"},
				[]string{"b"},
			},
			search: []string{"c", "d", "e"},
			want: [][]string{
				[]string{"c", "d", "e"},
				[]string{"c", "d", "e"},
			},
		},
		{
			input: [][]string{
				[]string{"a", "b", "c", "d"},
				[]string{"c"},
				[]string{"c", "d", "e"},
				[]string{"b"},
			},
			search: []string{"c"},
			want: [][]string{
				[]string{"c"},
			},
		},
		{
			input:  [][]string{[]string{"b"}},
			search: []string{"a"},
			want:   nil,
		},
		{
			input:  [][]string{[]string{"b"}},
			search: []string{"c"},
			want:   nil,
		},
		{
			input: [][]string{
				[]string{"a", "b", "c", "d"},
				[]string{"c", "#"},
				[]string{"c", "d", "e"},
				[]string{"b"},
				[]string{"#"},
			},
			search: []string{"c"},
			want: [][]string{
				[]string{"c", "#"},
				[]string{"#"},
			},
		},
	} {
		r := New()
		for _, topic := range test.input {
			r.Add(topic, topic)
		}
		result := r.Find(test.search)
		if len(result) != len(test.want) {
			t.Errorf("unexpected result %v != %v", result, test.want)
			continue
		}
		for i := range result {
			if !reflect.DeepEqual(result[i], test.want[i]) {
				t.Errorf("unexpected result[%d] %v != %v", i, result, test.want)
			}
		}
	}
}

func BenchmarkRouter(b *testing.B) {
	r := New()
	randomPath := func() []string {
		length := rand.Intn(4) + 2
		path := make([]string, length)
		for i := range path {
			path[i] = string('a'+rand.Intn(26)) + string('a'+rand.Intn(26))
		}
		if rand.Intn(10) == 0 {
			path[length-1] = "#"
		}
		return path
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
