package koanf_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/hcl"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

const (
	delim = "."

	mockDir  = "mock"
	mockJSON = mockDir + "/mock.json"
	mockYAML = mockDir + "/mock.yml"
	mockTOML = mockDir + "/mock.toml"
	mockHCL  = mockDir + "/mock.hcl"
	mockProp = mockDir + "/mock.prop"
)

// Ordered list of fields in the test confs.
var testKeys = []string{
	"bools",
	"duration",
	"empty",
	"intbools",
	"orphan",
	"parent1.boolmap.notok3",
	"parent1.boolmap.ok1",
	"parent1.boolmap.ok2",
	"parent1.child1.empty",
	"parent1.child1.grandchild1.ids",
	"parent1.child1.grandchild1.on",
	"parent1.child1.name",
	"parent1.child1.type",
	"parent1.floatmap.key1",
	"parent1.floatmap.key2",
	"parent1.floatmap.key3",
	"parent1.id",
	"parent1.intmap.key1",
	"parent1.intmap.key2",
	"parent1.intmap.key3",
	"parent1.name",
	"parent1.strmap.key1",
	"parent1.strmap.key2",
	"parent1.strmap.key3",
	"parent2.child2.empty",
	"parent2.child2.grandchild2.ids",
	"parent2.child2.grandchild2.on",
	"parent2.child2.name",
	"parent2.id",
	"parent2.name",
	"strbool",
	"strbools",
	"time",
	"type",
}

var testKeyMap = map[string][]string{
	"bools":                          {"bools"},
	"duration":                       {"duration"},
	"empty":                          {"empty"},
	"intbools":                       {"intbools"},
	"orphan":                         {"orphan"},
	"parent1":                        {"parent1"},
	"parent1.boolmap":                {"parent1", "boolmap"},
	"parent1.boolmap.notok3":         {"parent1", "boolmap", "notok3"},
	"parent1.boolmap.ok1":            {"parent1", "boolmap", "ok1"},
	"parent1.boolmap.ok2":            {"parent1", "boolmap", "ok2"},
	"parent1.child1":                 {"parent1", "child1"},
	"parent1.child1.empty":           {"parent1", "child1", "empty"},
	"parent1.child1.grandchild1":     {"parent1", "child1", "grandchild1"},
	"parent1.child1.grandchild1.ids": {"parent1", "child1", "grandchild1", "ids"},
	"parent1.child1.grandchild1.on":  {"parent1", "child1", "grandchild1", "on"},
	"parent1.child1.name":            {"parent1", "child1", "name"},
	"parent1.child1.type":            {"parent1", "child1", "type"},
	"parent1.floatmap":               {"parent1", "floatmap"},
	"parent1.floatmap.key1":          {"parent1", "floatmap", "key1"},
	"parent1.floatmap.key2":          {"parent1", "floatmap", "key2"},
	"parent1.floatmap.key3":          {"parent1", "floatmap", "key3"},
	"parent1.id":                     {"parent1", "id"},
	"parent1.intmap":                 {"parent1", "intmap"},
	"parent1.intmap.key1":            {"parent1", "intmap", "key1"},
	"parent1.intmap.key2":            {"parent1", "intmap", "key2"},
	"parent1.intmap.key3":            {"parent1", "intmap", "key3"},
	"parent1.name":                   {"parent1", "name"},
	"parent1.strmap":                 {"parent1", "strmap"},
	"parent1.strmap.key1":            {"parent1", "strmap", "key1"},
	"parent1.strmap.key2":            {"parent1", "strmap", "key2"},
	"parent1.strmap.key3":            {"parent1", "strmap", "key3"},
	"parent2":                        {"parent2"},
	"parent2.child2":                 {"parent2", "child2"},
	"parent2.child2.empty":           {"parent2", "child2", "empty"},
	"parent2.child2.grandchild2":     {"parent2", "child2", "grandchild2"},
	"parent2.child2.grandchild2.ids": {"parent2", "child2", "grandchild2", "ids"},
	"parent2.child2.grandchild2.on":  {"parent2", "child2", "grandchild2", "on"},
	"parent2.child2.name":            {"parent2", "child2", "name"},
	"parent2.id":                     {"parent2", "id"},
	"parent2.name":                   {"parent2", "name"},
	"strbool":                        {"strbool"},
	"strbools":                       {"strbools"},
	"time":                           {"time"},
	"type":                           {"type"},
}

// `parent1.child1.type` and `type` are excluded as their
// values vary between files.
var testAll = `bools -> [true false true]
duration -> 3s
empty -> map[]
intbools -> [1 0 1]
orphan -> [red blue orange]
parent1.boolmap.notok3 -> false
parent1.boolmap.ok1 -> true
parent1.boolmap.ok2 -> true
parent1.child1.empty -> map[]
parent1.child1.grandchild1.ids -> [1 2 3]
parent1.child1.grandchild1.on -> true
parent1.child1.name -> child1
parent1.floatmap.key1 -> 1.1
parent1.floatmap.key2 -> 1.2
parent1.floatmap.key3 -> 1.3
parent1.id -> 1234
parent1.intmap.key1 -> 1
parent1.intmap.key2 -> 1
parent1.intmap.key3 -> 1
parent1.name -> parent1
parent1.strmap.key1 -> val1
parent1.strmap.key2 -> val2
parent1.strmap.key3 -> val3
parent2.child2.empty -> map[]
parent2.child2.grandchild2.ids -> [4 5 6]
parent2.child2.grandchild2.on -> true
parent2.child2.name -> child2
parent2.id -> 5678
parent2.name -> parent2
strbool -> 1
strbools -> [1 t f]
time -> 2019-01-01`

var testParent2 = `child2.empty -> map[]
child2.grandchild2.ids -> [4 5 6]
child2.grandchild2.on -> true
child2.name -> child2
id -> 5678
name -> parent2`

type parentStruct struct {
	Name   string      `koanf:"name"`
	ID     int         `koanf:"id"`
	Child1 childStruct `koanf:"child1"`
}
type childStruct struct {
	Name        string            `koanf:"name"`
	Type        string            `koanf:"type"`
	Empty       map[string]string `koanf:"empty"`
	Grandchild1 grandchildStruct  `koanf:"grandchild1"`
}
type grandchildStruct struct {
	Ids []int `koanf:"ids"`
	On  bool  `koanf:"on"`
}
type testStruct struct {
	Type    string            `koanf:"type"`
	Empty   map[string]string `koanf:"empty"`
	Parent1 parentStruct      `koanf:"parent1"`
}

type testStructFlat struct {
	Type                        string            `koanf:"type"`
	Empty                       map[string]string `koanf:"empty"`
	Parent1Name                 string            `koanf:"parent1.name"`
	Parent1ID                   int               `koanf:"parent1.id"`
	Parent1Child1Name           string            `koanf:"parent1.child1.name"`
	Parent1Child1Type           string            `koanf:"parent1.child1.type"`
	Parent1Child1Empty          map[string]string `koanf:"parent1.child1.empty"`
	Parent1Child1Grandchild1IDs []int             `koanf:"parent1.child1.grandchild1.ids"`
	Parent1Child1Grandchild1On  bool              `koanf:"parent1.child1.grandchild1.on"`
}

type Case struct {
	koanf    *koanf.Koanf
	file     string
	parser   koanf.Parser
	typeName string
}

// Case instances to be used in multiple tests. These will not be mutated.
var cases = []Case{
	{koanf: koanf.New(delim), file: mockJSON, parser: json.Parser(), typeName: "json"},
	{koanf: koanf.New(delim), file: mockYAML, parser: yaml.Parser(), typeName: "yml"},
	{koanf: koanf.New(delim), file: mockTOML, parser: toml.Parser(), typeName: "toml"},
	{koanf: koanf.New(delim), file: mockHCL, parser: hcl.Parser(true), typeName: "hcl"},
}

func init() {
	// Preload 4 Koanf instances with their providers and config.
	if err := cases[0].koanf.Load(file.Provider(cases[0].file), json.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[1].koanf.Load(file.Provider(cases[1].file), yaml.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[2].koanf.Load(file.Provider(cases[2].file), toml.Parser()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
	if err := cases[3].koanf.Load(file.Provider(cases[3].file), hcl.Parser(true)); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}
}

func TestLoadFile(t *testing.T) {
	// Load a non-existent file.
	_, err := file.Provider("does-not-exist").ReadBytes()
	assert.NotNil(t, err, "no error for non-existent file")

	// Load a valid file.
	_, err = file.Provider(mockJSON).ReadBytes()
	assert.Nil(t, err, "error loading file")
}

func TestLoadFileAllKeys(t *testing.T) {
	assert := assert.New(t)
	re, _ := regexp.Compile("(.+?)?type \\-> (.*)\n")
	for _, c := range cases {
		// Check against testKeys.
		assert.Equal(testKeys, c.koanf.Keys(), fmt.Sprintf("loaded keys mismatch: %v", c.typeName))

		// Check against keyMap.
		assert.EqualValues(testKeyMap, c.koanf.KeyMap(), "keymap doesn't match")

		// Replace the "type" fields that varies across different files
		// to do a complete key -> value map match with testAll.
		s := strings.TrimSpace(re.ReplaceAllString(c.koanf.Sprint(), ""))
		assert.Equal(testAll, s, fmt.Sprintf("key -> value list mismatch: %v", c.typeName))
	}
}

func TestWatchFile(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	// Create a tmp config file.
	out, err := ioutil.TempFile("", "koanf_mock")
	if err != nil {
		log.Fatalf("error creating temp config file: %v", err)
	}
	out.Write([]byte(`{"parent": {"name": "name1"}}`))
	out.Close()

	// Load the new config and watch it for changes.
	f := file.Provider(out.Name())
	k.Load(f, json.Parser())

	// Watch for changes.
	changedName := ""
	f.Watch(func(event interface{}, err error) {
		// The File watcher always returns a nil `event`, which can
		// be ignored.
		assert.NoError(err, "watch file event error")

		if err != nil {
			return
		}
		// Reload the config.
		k.Load(f, json.Parser())
		changedName = k.String("parent.name")
	})

	// Wait a second and change the file.
	time.Sleep(1 * time.Second)
	ioutil.WriteFile(out.Name(), []byte(`{"parent": {"name": "name2"}}`), 0644)
	time.Sleep(1 * time.Second)

	assert.Equal("name2", changedName, "file watch reload didn't change config")
}

func TestWatchFileSymlink(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)

	// Create a symlink.
	symPath := filepath.Join(os.TempDir(), "koanf_test_symlink")
	os.Remove(symPath)
	symPath2 := filepath.Join(os.TempDir(), "koanf_test_symlink2")
	os.Remove(symPath)

	wd, err := os.Getwd()
	assert.NoError(err, "error getting working dir")

	jsonFile := filepath.Join(wd, mockJSON)
	yamlFile := filepath.Join(wd, mockYAML)

	// Create a symlink to the JSON file which will be swapped out later.
	assert.NoError(os.Symlink(jsonFile, symPath), "error creating symlink")

	// Load the symlink (to the JSON) file.
	f := file.Provider(symPath)
	k.Load(f, json.Parser())

	// Watch for changes.
	changedType := ""
	f.Watch(func(event interface{}, err error) {
		// The File watcher always returns a nil `event`, which can
		// be ignored.
		assert.NoError(err, "watch file event error")

		if err != nil {
			return
		}
		// Reload the config.
		k.Load(f, yaml.Parser())
		changedType = k.String("type")
	})

	// Wait a second and swap the symlink target from the JSON file to the YAML file.
	// Create a temp symlink to the YAML file and rename the old symlink to the new
	// symlink. We do this to avoid removing the symlink and triggering a REMOVE event.
	time.Sleep(1 * time.Second)
	assert.NoError(os.Symlink(yamlFile, symPath2), "error creating temp symlink")
	assert.NoError(os.Rename(symPath2, symPath), "error creating temp symlink")
	time.Sleep(1 * time.Second)

	assert.Equal("yml", changedType, "symlink watch reload didn't change config")
}

func TestLoadMerge(t *testing.T) {
	var (
		assert = assert.New(t)
		// Load several types into a fresh Koanf instance.
		k = koanf.New(delim)
	)
	for _, c := range cases {
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Check against testKeys.
		assert.Equal(testKeys, k.Keys(), fmt.Sprintf("loaded keys don't match in: %v", c.file))

		// The 'type' fields in different file types have different values.
		// As each subsequent file is loaded, the previous value should be overridden.
		assert.Equal(c.typeName, k.String("type"), "types don't match")
		assert.Equal(c.typeName, k.String("parent1.child1.type"), "types don't match")
	}

	// Load env provider and override value.
	os.Setenv("PREFIX_PARENT1.CHILD1.TYPE", "env")
	err := k.Load(env.Provider("PREFIX_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "prefix_", "", -1)
	}), nil)

	assert.Nil(err, "error loading env")
	assert.Equal("env", k.String("parent1.child1.type"), "types don't match")

	// Override with the confmap provider.
	k.Load(confmap.Provider(map[string]interface{}{
		"parent1.child1.type": "confmap",
		"type":                "confmap",
	}, "."), nil)
	assert.Equal("confmap", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("confmap", k.String("type"), "types don't match")

	// Override with the rawbytes provider.
	assert.Nil(k.Load(rawbytes.Provider([]byte(`{"type": "rawbytes", "parent1": {"child1": {"type": "rawbytes"}}}`)), json.Parser()),
		"error loading raw bytes")
	assert.Equal("rawbytes", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("rawbytes", k.String("type"), "types don't match")
}

func TestFlags(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	k2 := k.Copy()

	// Override with the posflag provider.
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)

	// Key that exists in the loaded conf. Should overwrite with Set().
	f.String("parent1.child1.type", "flag", "")
	f.Set("parent1.child1.type", "flag")
	f.StringSlice("stringslice", []string{"a", "b", "c"}, "")
	f.IntSlice("intslice", []int{1, 2, 3}, "")

	// Key that doesn't exist in the loaded file conf. Should merge the default value.
	f.String("flagkey", "flag", "")

	// Key that exists in the loadd conf but no Set(). Default value shouldn't be merged.
	f.String("parent1.name", "flag", "")

	// Initialize the provider with the Koanf instance passed where default values
	// will merge if the keys are not present in the conf map.
	assert.Nil(k.Load(posflag.Provider(f, ".", k), nil), "error loading posflag")
	assert.Equal("flag", k.String("parent1.child1.type"), "types don't match")
	assert.Equal("flag", k.String("flagkey"), "value doesn't match")
	assert.NotEqual("flag", k.String("parent1.name"), "value doesn't match")
	assert.Equal([]string{"a", "b", "c"}, k.Strings("stringslice"), "value doesn't match")
	assert.Equal([]int{1, 2, 3}, k.Ints("intslice"), "value doesn't match")

	// Test without passing the Koanf instance where default values will not merge.
	assert.Nil(k2.Load(posflag.Provider(f, ".", nil), nil), "error loading posflag")
	assert.Equal("flag", k2.String("parent1.child1.type"), "types don't match")
	assert.Equal("", k2.String("flagkey"), "value doesn't match")
	assert.NotEqual("", k2.String("parent1.name"), "value doesn't match")

	// Override with the flag provider.
	bf := flag.NewFlagSet("test", flag.ContinueOnError)
	bf.String("parent1.child1.type", "flag", "")
	bf.Set("parent1.child1.type", "basicflag")
	assert.Nil(k.Load(basicflag.Provider(bf, "."), nil), "error loading basicflag")
	assert.Equal("basicflag", k.String("parent1.child1.type"), "types don't match")

}

func TestConfMapValues(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c1  = k.All()
		ra1 = k.Raw()
		r1  = k.Cut("parent2").Raw()
	)

	k = koanf.New(delim)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()), "error loading file")
	var (
		c2  = k.All()
		ra2 = k.Raw()
		r2  = k.Cut("parent2").Raw()
	)

	assert.EqualValues(c1, c2, "conf map mismatch")
	assert.EqualValues(ra1, ra2, "conf map mismatch")
	assert.EqualValues(r1, r2, "conf map mismatch")
}

func TestCutCopy(t *testing.T) {
	// Instance 1.
	var (
		assert = assert.New(t)
		k1     = koanf.New(delim)
	)
	assert.Nil(k1.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp1   = k1.Copy()
		cut1  = k1.Cut("")
		cutp1 = k1.Cut("parent2")
	)

	// Instance 2.
	k2 := koanf.New(delim)
	assert.Nil(k2.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")
	var (
		cp2   = k2.Copy()
		cut2  = k2.Cut("")
		cutp2 = k2.Cut("parent2")
	)

	assert.EqualValues(cp1.All(), cp2.All(), "conf map mismatch")
	assert.EqualValues(cut1.All(), cut2.All(), "conf map mismatch")
	assert.EqualValues(cutp1.All(), cutp2.All(), "conf map mismatch")
	assert.Equal(testParent2, strings.TrimSpace(cutp1.Sprint()), "conf map mismatch")
	assert.Equal(strings.TrimSpace(cutp1.Sprint()), strings.TrimSpace(cutp2.Sprint()), "conf map mismatch")

	// Cut a single field with no children. Should return empty conf maps.
	assert.Equal(k1.Cut("type").Keys(), k2.Cut("type").Keys(), "single field cut mismatch")
	assert.Equal(k1.Cut("xxxx").Keys(), k2.Cut("xxxx").Keys(), "single field cut mismatch")
	assert.Equal(len(k1.Cut("type").Raw()), 0, "non-map cut returned items")
}

func TestMerge(t *testing.T) {
	var (
		assert = assert.New(t)
		k      = koanf.New(delim)
	)
	assert.Nil(k.Load(file.Provider(mockJSON), json.Parser()),
		"error loading file")

	// Make two different cuts that'll have different confmaps.
	var (
		cut1 = k.Cut("parent1")
		cut2 = k.Cut("parent2")
	)
	assert.NotEqual(cut1.All(), cut2.All(), "different cuts incorrectly match")
	assert.NotEqual(cut1.Sprint(), cut2.Sprint(), "different cuts incorrectly match")

	// Create an empty Koanf instance.
	k2 := koanf.New(delim)

	// Merge cut1 into it and check if they match.
	k2.Merge(cut1)
	assert.Equal(cut1.All(), k2.All(), "conf map mismatch")
}

func TestUnmarshal(t *testing.T) {
	assert := assert.New(t)

	// Expected unmarshalled structure.
	real := testStruct{
		Type:  "json",
		Empty: make(map[string]string),
		Parent1: parentStruct{
			Name: "parent1",
			ID:   1234,
			Child1: childStruct{
				Name:  "child1",
				Type:  "json",
				Empty: make(map[string]string),
				Grandchild1: grandchildStruct{
					Ids: []int{1, 2, 3},
					On:  true,
				},
			},
		},
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		var (
			k  = koanf.New(delim)
			ts testStruct
		)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		assert.Nil(k.Unmarshal("", &ts), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")

		// Unmarshal with config.
		ts = testStruct{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf"}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1.Child1.Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestUnmarshalFlat(t *testing.T) {
	assert := assert.New(t)

	// Expected unmarshalled structure.
	real := testStructFlat{
		Type:                        "json",
		Empty:                       make(map[string]string),
		Parent1Name:                 "parent1",
		Parent1ID:                   1234,
		Parent1Child1Name:           "child1",
		Parent1Child1Type:           "json",
		Parent1Child1Empty:          make(map[string]string),
		Parent1Child1Grandchild1IDs: []int{1, 2, 3},
		Parent1Child1Grandchild1On:  true,
	}

	// Unmarshal and check all parsers.
	for _, c := range cases {
		k := koanf.New(delim)
		assert.Nil(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))
		ts := testStructFlat{}
		assert.Nil(k.UnmarshalWithConf("", &ts, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true}), "unmarshal failed")
		real.Type = c.typeName
		real.Parent1Child1Type = c.typeName
		assert.Equal(real, ts, "unmarshalled structs don't match")
	}
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	for _, c := range cases {
		// HCL does not support marshalling.
		if c.typeName == "hcl" {
			continue
		}

		// Load config.
		k := koanf.New(delim)
		assert.NoError(k.Load(file.Provider(c.file), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Serialize / marshal into raw bytes using the parser.
		b, err := k.Marshal(c.parser)
		assert.NoError(err, "error marshalling")

		// Reload raw serialize bytes into a new koanf instance.
		k = koanf.New(delim)
		assert.NoError(k.Load(rawbytes.Provider(b), c.parser),
			fmt.Sprintf("error loading: %v", c.file))

		// Check if values are intact.
		assert.Equal(float64(1234), c.koanf.MustFloat64("parent1.id"))
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.MustStrings("orphan"))
		assert.Equal([]int64{1, 2, 3}, c.koanf.MustInt64s("parent1.child1.grandchild1.ids"))
	}
}

func TestGetExists(t *testing.T) {
	assert := assert.New(t)

	type exCase struct {
		path   string
		exists bool
	}
	exCases := []exCase{
		{"xxxxx", false},
		{"parent1.child2", false},
		{"child1", false},
		{"child2", false},
		{"type", true},
		{"parent1", true},
		{"parent2", true},
		{"parent1.name", true},
		{"parent2.name", true},
		{"parent1.child1", true},
		{"parent2.child2", true},
		{"parent1.child1.grandchild1", true},
		{"parent1.child1.grandchild1.on", true},
		{"parent2.child2.grandchild2.on", true},
	}
	for _, c := range exCases {
		assert.Equal(c.exists, cases[0].koanf.Exists(c.path),
			fmt.Sprintf("path resolution failed: %s", c.path))
		assert.Equal(c.exists, cases[0].koanf.Get(c.path) != nil,
			fmt.Sprintf("path resolution failed: %s", c.path))
	}
}

func TestGetTypes(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		assert.Equal(nil, c.koanf.Get("xxx"))
		assert.Equal(make(map[string]interface{}), c.koanf.Get("empty"))

		// Int.
		assert.Equal(int64(0), c.koanf.Int64("xxxx"))
		assert.Equal(int64(1234), c.koanf.Int64("parent1.id"))

		assert.Equal(int(0), c.koanf.Int("xxxx"))
		assert.Equal(int(1234), c.koanf.Int("parent1.id"))

		assert.Equal([]int64{}, c.koanf.Int64s("xxxx"))
		assert.Equal([]int64{1, 2, 3}, c.koanf.Int64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.intmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("parent1.boolmap"))
		assert.Equal(map[string]int64{}, c.koanf.Int64Map("xxxx"))
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Int64Map("parent1.floatmap"))

		assert.Equal([]int{1, 2, 3}, c.koanf.Ints("parent1.child1.grandchild1.ids"))
		assert.Equal([]int{}, c.koanf.Ints("xxxx"))

		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.IntMap("parent1.intmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("parent1.boolmap"))
		assert.Equal(map[string]int{}, c.koanf.IntMap("xxxx"))

		// Float.
		assert.Equal(float64(0), c.koanf.Float64("xxx"))
		assert.Equal(float64(1234), c.koanf.Float64("parent1.id"))

		assert.Equal([]float64{}, c.koanf.Float64s("xxxx"))
		assert.Equal([]float64{1, 2, 3}, c.koanf.Float64s("parent1.child1.grandchild1.ids"))

		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.Float64Map("parent1.intmap"))
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.Float64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("parent1.boolmap"))
		assert.Equal(map[string]float64{}, c.koanf.Float64Map("xxxx"))

		// String and bytes.
		assert.Equal([]byte{}, c.koanf.Bytes("xxxx"))
		assert.Equal([]byte("parent1"), c.koanf.Bytes("parent1.name"))

		assert.Equal("", c.koanf.String("xxxx"))
		assert.Equal("parent1", c.koanf.String("parent1.name"))

		assert.Equal([]string{}, c.koanf.Strings("xxxx"))
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.Strings("orphan"))

		assert.Equal(map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}, c.koanf.StringMap("parent1.strmap"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("xxxx"))
		assert.Equal(map[string]string{}, c.koanf.StringMap("parent1.intmap"))

		// Bools.
		assert.Equal(false, c.koanf.Bool("xxxx"))
		assert.Equal(false, c.koanf.Bool("type"))
		assert.Equal(true, c.koanf.Bool("parent1.child1.grandchild1.on"))
		assert.Equal(true, c.koanf.Bool("strbool"))

		assert.Equal([]bool{}, c.koanf.Bools("xxxx"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("bools"))
		assert.Equal([]bool{true, false, true}, c.koanf.Bools("intbools"))
		assert.Equal([]bool{true, true, false}, c.koanf.Bools("strbools"))

		assert.Equal(map[string]bool{"ok1": true, "ok2": true, "notok3": false}, c.koanf.BoolMap("parent1.boolmap"))
		assert.Equal(map[string]bool{"key1": true, "key2": true, "key3": true}, c.koanf.BoolMap("parent1.intmap"))
		assert.Equal(map[string]bool{}, c.koanf.BoolMap("xxxx"))

		// Others.
		assert.Equal(time.Duration(1234), c.koanf.Duration("parent1.id"))
		assert.Equal(time.Duration(0), c.koanf.Duration("xxxx"))
		assert.Equal(time.Second*3, c.koanf.Duration("duration"))

		assert.Equal(time.Time{}, c.koanf.Time("xxxx", "2006-01-02"))
		assert.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.Time("time", "2006-01-02"))

		assert.Equal([]string{}, c.koanf.MapKeys("xxxx"), "map keys mismatch")
		assert.Equal([]string{"bools", "duration", "empty", "intbools", "orphan", "parent1", "parent2", "strbool", "strbools", "time", "type"},
			c.koanf.MapKeys(""), "map keys mismatch")
		assert.Equal([]string{"key1", "key2", "key3"}, c.koanf.MapKeys("parent1.strmap"), "map keys mismatch")

		// Attempt to parse int=1234 as a Unix timestamp.
		assert.Equal(time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.Time("parent1.id", "").UTC())
	}
}

func TestMustGetTypes(t *testing.T) {
	assert := assert.New(t)
	for _, c := range cases {
		// Int.
		assert.Panics(func() { c.koanf.MustInt64("xxxx") })
		assert.Equal(int64(1234), c.koanf.MustInt64("parent1.id"))

		assert.Panics(func() { c.koanf.MustInt("xxxx") })
		assert.Equal(int(1234), c.koanf.MustInt("parent1.id"))

		assert.Panics(func() { c.koanf.MustInt64s("xxxx") })
		assert.Equal([]int64{1, 2, 3}, c.koanf.MustInt64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustInt64Map("xxxx") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.intmap"))

		assert.Panics(func() { c.koanf.MustInt64Map("parent1.boolmap") })
		assert.Equal(map[string]int64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustInt64Map("parent1.floatmap"))

		assert.Panics(func() { c.koanf.MustInts("xxxx") })
		assert.Equal([]int{1, 2, 3}, c.koanf.MustInts("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustIntMap("xxxx") })
		assert.Panics(func() { c.koanf.MustIntMap("parent1.boolmap") })
		assert.Equal(map[string]int{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustIntMap("parent1.intmap"))

		// Float.
		assert.Panics(func() { c.koanf.MustInts("xxxx") })
		assert.Equal(float64(1234), c.koanf.MustFloat64("parent1.id"))

		assert.Panics(func() { c.koanf.MustFloat64s("xxxx") })
		assert.Equal([]float64{1, 2, 3}, c.koanf.MustFloat64s("parent1.child1.grandchild1.ids"))

		assert.Panics(func() { c.koanf.MustFloat64Map("xxxx") })
		assert.Panics(func() { c.koanf.MustFloat64Map("parent1.boolmap") })
		assert.Equal(map[string]float64{"key1": 1.1, "key2": 1.2, "key3": 1.3}, c.koanf.MustFloat64Map("parent1.floatmap"))
		assert.Equal(map[string]float64{"key1": 1, "key2": 1, "key3": 1}, c.koanf.MustFloat64Map("parent1.intmap"))

		// String and bytes.
		assert.Panics(func() { c.koanf.MustBytes("xxxx") })
		assert.Equal([]byte("parent1"), c.koanf.MustBytes("parent1.name"))

		assert.Panics(func() { c.koanf.MustString("xxxx") })
		assert.Equal("parent1", c.koanf.MustString("parent1.name"))

		assert.Panics(func() { c.koanf.MustStrings("xxxx") })
		assert.Equal([]string{"red", "blue", "orange"}, c.koanf.MustStrings("orphan"))

		assert.Panics(func() { c.koanf.MustStringMap("xxxx") })
		assert.Panics(func() { c.koanf.MustStringMap("parent1.intmap") })
		assert.Equal(map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}, c.koanf.MustStringMap("parent1.strmap"))

		// // Bools.
		assert.Panics(func() { c.koanf.MustBools("xxxx") })
		assert.Equal([]bool{true, false, true}, c.koanf.MustBools("bools"))
		assert.Equal([]bool{true, false, true}, c.koanf.MustBools("intbools"))
		assert.Equal([]bool{true, true, false}, c.koanf.MustBools("strbools"))

		assert.Panics(func() { c.koanf.MustBoolMap("xxxx") })
		assert.Equal(map[string]bool{"ok1": true, "ok2": true, "notok3": false}, c.koanf.MustBoolMap("parent1.boolmap"))
		assert.Equal(map[string]bool{"key1": true, "key2": true, "key3": true}, c.koanf.MustBoolMap("parent1.intmap"))

		// Others.
		assert.Panics(func() { c.koanf.MustDuration("xxxx") })
		assert.Equal(time.Duration(1234), c.koanf.MustDuration("parent1.id"))
		assert.Equal(time.Second*3, c.koanf.MustDuration("duration"))

		assert.Panics(func() { c.koanf.MustTime("xxxx", "2006-01-02") })
		assert.Equal(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), c.koanf.MustTime("time", "2006-01-02"))

		// // Attempt to parse int=1234 as a Unix timestamp.
		assert.Panics(func() { c.koanf.MustTime("time", "2006") })
		assert.Equal(time.Date(1970, 1, 1, 0, 20, 34, 0, time.UTC), c.koanf.MustTime("parent1.id", "").UTC())
	}
}
