package tsm1

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/influxdb/influxdb/models"
	"github.com/influxdb/influxdb/tsdb"
)

// Ensure an engine containing cached values responds correctly to queries.
func TestEngine_QueryCache_Ascending(t *testing.T) {
	// Generate temporary file.
	f, _ := ioutil.TempFile("", "tsm")
	f.Close()
	os.Remove(f.Name())
	walPath := filepath.Join(f.Name(), "wal")
	os.MkdirAll(walPath, 0777)
	defer os.RemoveAll(f.Name())

	// Create a few points.
	p1 := parsePoint("cpu,host=A value=1.1 1000000000")
	p2 := parsePoint("cpu,host=A value=1.2 2000000000")
	p3 := parsePoint("cpu,host=A value=1.3 3000000000")

	// Write those points to the engine.
	e := NewEngine(f.Name(), walPath, tsdb.NewEngineOptions())
	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}
	if err := e.WritePoints([]models.Point{p1, p2, p3}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	// Start a query transactions and get a cursor.
	tx := tx{engine: e.(*Engine)}
	ascCursor := tx.Cursor("cpu,host=A", []string{"value"}, nil, true)

	k, v := ascCursor.SeekTo(1)
	if k != 1000000000 {
		t.Fatalf("failed to seek to before first key: %v %v", k, v)
	}

	k, v = ascCursor.SeekTo(1000000000)
	if k != 1000000000 {
		t.Fatalf("failed to seek to first key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != 2000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != 3000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != -1 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.SeekTo(4000000000)
	if k != -1 {
		t.Fatalf("failed to seek to past last key: %v %v", k, v)
	}
}

// Ensure an engine containing cached values responds correctly to queries.
func TestEngine_QueryTSM_Ascending(t *testing.T) {
	fs := NewFileStore("")

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(1, 0), 1.0)}},
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(2, 0), 2.0)}},
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(3, 0), 3.0)}},
	}

	files, err := newFiles(data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Start a query transactions and get a cursor.
	ascCursor := devCursor{
		tsmKeyCursor: fs.KeyCursor("cpu,host=A#!~#value"),
		series:       "cpu,host=A",
		fields:       []string{"value"},
		ascending:    true,
	}

	k, v := ascCursor.SeekTo(1)
	if k != 1000000000 {
		t.Fatalf("failed to seek to before first key: %v %v", k, v)
	}

	k, v = ascCursor.SeekTo(1000000000)
	if k != 1000000000 {
		t.Fatalf("failed to seek to first key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != 2000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != 3000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.Next()
	if k != -1 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = ascCursor.SeekTo(4000000000)
	if k != -1 {
		t.Fatalf("failed to seek to past last key: %v %v", k, v)
	}
}

// Ensure an engine containing cached values responds correctly to queries.
func TestEngine_QueryCache_Descending(t *testing.T) {
	// Generate temporary file.
	f, _ := ioutil.TempFile("", "tsm")
	f.Close()
	os.Remove(f.Name())
	walPath := filepath.Join(f.Name(), "wal")
	os.MkdirAll(walPath, 0777)
	defer os.RemoveAll(f.Name())

	// Create a few points.
	p1 := parsePoint("cpu,host=A value=1.1 1000000000")
	p2 := parsePoint("cpu,host=A value=1.2 2000000000")
	p3 := parsePoint("cpu,host=A value=1.3 3000000000")

	// Write those points to the engine.
	e := NewEngine(f.Name(), walPath, tsdb.NewEngineOptions())
	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}
	if err := e.WritePoints([]models.Point{p1, p2, p3}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	// Start a query transactions and get a cursor.
	tx := tx{engine: e.(*Engine)}
	descCursor := tx.Cursor("cpu,host=A", []string{"value"}, nil, false)

	k, v := descCursor.SeekTo(4000000000)
	if k != 3000000000 {
		t.Fatalf("failed to seek to before last key: %v %v", k, v)
	}

	k, v = descCursor.Next()
	if k != 2000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = descCursor.SeekTo(1)
	if k != -1 {
		t.Fatalf("failed to seek to after first key: %v %v", k, v)
	}
}

// Ensure an engine containing cached values responds correctly to queries.
func TestEngine_QueryTSM_Descending(t *testing.T) {
	fs := NewFileStore("")

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(1, 0), 1.0)}},
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(2, 0), 2.0)}},
		keyValues{"cpu,host=A#!~#value", []Value{NewValue(time.Unix(3, 0), 3.0)}},
	}

	files, err := newFiles(data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Start a query transactions and get a cursor.
	descCursor := devCursor{
		tsmKeyCursor: fs.KeyCursor("cpu,host=A#!~#value"),
		series:       "cpu,host=A",
		fields:       []string{"value"},
		ascending:    false,
	}

	k, v := descCursor.SeekTo(4000000000)
	if k != 3000000000 {
		t.Fatalf("failed to seek to before last key: %v %v", k, v)
	}

	k, v = descCursor.Next()
	if k != 2000000000 {
		t.Fatalf("failed to get next key: %v %v", k, v)
	}

	k, v = descCursor.SeekTo(1)
	if k != -1 {
		t.Fatalf("failed to seek to after first key: %v %v", k, v)
	}
}

func TestEngine_LoadMetadataIndex(t *testing.T) {
	// Generate temporary file.
	f, _ := ioutil.TempFile("", "tsm")
	f.Close()
	os.Remove(f.Name())
	walPath := filepath.Join(f.Name(), "wal")
	os.MkdirAll(walPath, 0777)
	defer os.RemoveAll(f.Name())

	// Create a few points.
	p1 := parsePoint("cpu,host=A value=1.1 1000000000")
	p2 := parsePoint("cpu,host=B value=1.2 2000000000")

	// Write those points to the engine.
	e := NewEngine(f.Name(), walPath, tsdb.NewEngineOptions()).(*Engine)
	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}
	if err := e.WritePoints([]models.Point{p1}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	// ensure we can close and load index from the WAL
	if err := e.Close(); err != nil {
		t.Fatalf("error closing: %s", err.Error())
	}
	if err := e.Open(); err != nil {
		t.Fatalf("error opening: %s", err.Error())
	}

	// Load metadata index.
	index := tsdb.NewDatabaseIndex()
	if err := e.LoadMetadataIndex(nil, index, make(map[string]*tsdb.MeasurementFields)); err != nil {
		t.Fatal(err)
	}

	// Verify index is correct.
	if m := index.Measurement("cpu"); m == nil {
		t.Fatal("measurement not found")
	} else if s := m.SeriesByID(1); s.Key != "cpu,host=A" || !reflect.DeepEqual(s.Tags, map[string]string{"host": "A"}) {
		t.Fatalf("unexpected series: %q / %#v", s.Key, s.Tags)
	}

	// write the snapshot, ensure we can close and load index from TSM
	if err := e.WriteSnapshot(); err != nil {
		t.Fatalf("error writing snapshot: %s", err.Error())
	}

	// ensure we can close and load index from the WAL
	if err := e.Close(); err != nil {
		t.Fatalf("error closing: %s", err.Error())
	}
	if err := e.Open(); err != nil {
		t.Fatalf("error opening: %s", err.Error())
	}

	// Load metadata index.
	index = tsdb.NewDatabaseIndex()
	if err := e.LoadMetadataIndex(nil, index, make(map[string]*tsdb.MeasurementFields)); err != nil {
		t.Fatal(err)
	}

	// Verify index is correct.
	if m := index.Measurement("cpu"); m == nil {
		t.Fatal("measurement not found")
	} else if s := m.SeriesByID(1); s.Key != "cpu,host=A" || !reflect.DeepEqual(s.Tags, map[string]string{"host": "A"}) {
		t.Fatalf("unexpected series: %q / %#v", s.Key, s.Tags)
	}

	// write a new point and ensure we can close and load index from TSM and WAL
	if err := e.WritePoints([]models.Point{p2}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	// ensure we can close and load index from the TSM and WAL
	if err := e.Close(); err != nil {
		t.Fatalf("error closing: %s", err.Error())
	}
	if err := e.Open(); err != nil {
		t.Fatalf("error opening: %s", err.Error())
	}

	// Load metadata index.
	index = tsdb.NewDatabaseIndex()
	if err := e.LoadMetadataIndex(nil, index, make(map[string]*tsdb.MeasurementFields)); err != nil {
		t.Fatal(err)
	}

	// Verify index is correct.
	if m := index.Measurement("cpu"); m == nil {
		t.Fatal("measurement not found")
	} else if s := m.SeriesByID(1); s.Key != "cpu,host=A" || !reflect.DeepEqual(s.Tags, map[string]string{"host": "A"}) {
		t.Fatalf("unexpected series: %q / %#v", s.Key, s.Tags)
	} else if s := m.SeriesByID(2); s.Key != "cpu,host=B" || !reflect.DeepEqual(s.Tags, map[string]string{"host": "B"}) {
		t.Fatalf("unexpected series: %q / %#v", s.Key, s.Tags)
	}
}

// Ensure that deletes only sent to the WAL will clear out the data from the cache on restart
func TestEngine_DeleteWALLoadMetadata(t *testing.T) {
	// Generate temporary file.
	f, _ := ioutil.TempFile("", "tsm")
	f.Close()
	os.Remove(f.Name())
	walPath := filepath.Join(f.Name(), "wal")
	os.MkdirAll(walPath, 0777)
	defer os.RemoveAll(f.Name())

	// Create a few points.
	p1 := parsePoint("cpu,host=A value=1.1 1000000000")
	p2 := parsePoint("cpu,host=B value=1.2 2000000000")

	// Write those points to the engine.
	e := NewEngine(f.Name(), walPath, tsdb.NewEngineOptions()).(*Engine)
	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}
	if err := e.WritePoints([]models.Point{p1, p2}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}
	if err := e.DeleteSeries([]string{"cpu,host=A"}); err != nil {
		t.Fatalf("failed to delete series: %s", err.Error())
	}

	// ensure we can close and load index from the WAL
	if err := e.Close(); err != nil {
		t.Fatalf("error closing: %s", err.Error())
	}

	e = NewEngine(f.Name(), walPath, tsdb.NewEngineOptions()).(*Engine)
	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}

	if exp, got := 0, len(e.Cache.Values(SeriesFieldKey("cpu,host=A", "value"))); exp != got {
		t.Fatalf("unexpected number of values: got: %d. exp: %d", got, exp)
	}

	if exp, got := 1, len(e.Cache.Values(SeriesFieldKey("cpu,host=B", "value"))); exp != got {
		t.Fatalf("unexpected number of values: got: %d. exp: %d", got, exp)
	}
}

// Ensure that the engine will backup any TSM files created since the passed in time
func TestEngine_Backup(t *testing.T) {
	// Generate temporary file.
	f, _ := ioutil.TempFile("", "tsm")
	f.Close()
	os.Remove(f.Name())
	walPath := filepath.Join(f.Name(), "wal")
	os.MkdirAll(walPath, 0777)
	defer os.RemoveAll(f.Name())

	// Create a few points.
	p1 := parsePoint("cpu,host=A value=1.1 1000000000")
	p2 := parsePoint("cpu,host=B value=1.2 2000000000")
	p3 := parsePoint("cpu,host=C value=1.3 3000000000")

	// Write those points to the engine.
	e := NewEngine(f.Name(), walPath, tsdb.NewEngineOptions()).(*Engine)

	// mock the planner so compactions don't run during the test
	e.CompactionPlan = &mockPlanner{}

	if err := e.Open(); err != nil {
		t.Fatalf("failed to open tsm1 engine: %s", err.Error())
	}

	if err := e.WritePoints([]models.Point{p1}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}
	if err := e.WriteSnapshot(); err != nil {
		t.Fatalf("failed to snapshot: %s", err.Error())
	}

	if err := e.WritePoints([]models.Point{p2}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	b := bytes.NewBuffer(nil)
	if err := e.Backup(b, "", time.Unix(0, 0)); err != nil {
		t.Fatalf("failed to backup: %s", err.Error())
	}

	tr := tar.NewReader(b)
	if len(e.FileStore.files) != 2 {
		t.Fatalf("file count wrong: exp: %d, got: %d", 2, len(e.FileStore.files))
	}

	for _, f := range e.FileStore.files {
		th, err := tr.Next()
		if err != nil {
			t.Fatalf("failed reading header: %s", err.Error())
		}
		if !strings.Contains(f.Path(), th.Name) || th.Name == "" {
			t.Fatalf("file name doesn't match:\n\tgot: %s\n\texp: %s", th.Name, f.Path())
		}
	}

	lastBackup := time.Now()

	// we have to sleep for a second because last modified times only have second level precision.
	// so this test won't work properly unless the file is at least a second past the last one
	time.Sleep(time.Second)

	if err := e.WritePoints([]models.Point{p3}, nil, nil); err != nil {
		t.Fatalf("failed to write points: %s", err.Error())
	}

	b = bytes.NewBuffer(nil)
	if err := e.Backup(b, "", lastBackup); err != nil {
		t.Fatalf("failed to backup: %s", err.Error())
	}

	tr = tar.NewReader(b)
	th, err := tr.Next()
	if err != nil {
		t.Fatalf("error getting next tar header: %s", err.Error())
	}

	mostRecentFile := e.FileStore.files[len(e.FileStore.files)-1].Path()
	if !strings.Contains(mostRecentFile, th.Name) || th.Name == "" {
		t.Fatalf("file name doesn't match:\n\tgot: %s\n\texp: %s", th.Name, mostRecentFile)
	}
}

func parsePoints(buf string) []models.Point {
	points, err := models.ParsePointsString(buf)
	if err != nil {
		panic(fmt.Sprintf("couldn't parse points: %s", err.Error()))
	}
	return points
}

func parsePoint(buf string) models.Point {
	return parsePoints(buf)[0]
}

func newFiles(values ...keyValues) ([]TSMFile, error) {
	var files []TSMFile

	for _, v := range values {
		var b bytes.Buffer
		w, err := NewTSMWriter(&b)
		if err != nil {
			return nil, err
		}

		if err := w.Write(v.key, v.values); err != nil {
			return nil, err
		}

		if err := w.WriteIndex(); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}

		r, err := NewTSMReader(bytes.NewReader(b.Bytes()))
		if err != nil {
			return nil, err
		}
		files = append(files, r)
	}
	return files, nil
}

type keyValues struct {
	key    string
	values []Value
}

type mockPlanner struct{}

func (m *mockPlanner) Plan(lastWrite time.Time) []CompactionGroup { return nil }
func (m *mockPlanner) PlanLevel(level int) []CompactionGroup      { return nil }
