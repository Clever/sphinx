package kayvee

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

type Tests struct {
	Version        string     `json:"version"`
	FormatTests    []TestSpec `json:"format"`
	FormatLogTests []TestSpec `json:"formatLog"`
}

type TestSpec struct {
	Title  string                 `json:"title"`
	Input  map[string]interface{} `json:"input"`
	Output string                 `json:"output"`
}

type keyVal map[string]interface{}

// takes two strings (which are assumed to be JSON)
func compareJSONStrings(t *testing.T, expected string, actual string) {
	actualJSON := map[string]interface{}{}
	expectedJSON := map[string]interface{}{}
	err := json.Unmarshal([]byte(actual), &actualJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `actual`:", actual))
	}
	err = json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `expected`:", expected))
	}

	assert.Equal(t, expectedJSON, actualJSON)
}

func Test_KayveeSpecs(t *testing.T) {
	file, err := ioutil.ReadFile("tests.json")
	assert.NoError(t, err, "failed to open test specs (tests.json)")
	var tests Tests
	json.Unmarshal(file, &tests)
	t.Logf("spec (tests.json) version: %s\n", string(tests.Version))

	for _, spec := range tests.FormatTests {
		expected := spec.Output
		actual := Format(spec.Input["data"].(map[string]interface{}))

		compareJSONStrings(t, expected, actual)
	}

	for _, spec := range tests.FormatLogTests {
		expected := spec.Output

		// Ensure correct type is passed to FormatLog, even if value undefined in tests.json
		source, _ := spec.Input["source"].(string)
		level, _ := spec.Input["level"].(string)
		title, _ := spec.Input["title"].(string)
		data, _ := spec.Input["data"].(map[string]interface{})
		actual := FormatLog(source, level, title, data)

		compareJSONStrings(t, expected, actual)
	}

}
