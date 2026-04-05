package gswag

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/types"
)

// HaveStatus succeeds when the RecordedResponse has the expected HTTP status code.
func HaveStatus(expected int) types.GomegaMatcher {
	return &haveStatusMatcher{expected: expected}
}

type haveStatusMatcher struct{ expected int }

func (m *haveStatusMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}
	return res.StatusCode == m.expected, nil
}

func (m *haveStatusMatcher) FailureMessage(actual interface{}) string {
	res, _ := toRecordedResponse(actual)
	return fmt.Sprintf("Expected status %d but got %d\nBody: %s", m.expected, res.StatusCode, string(res.BodyBytes))
}

func (m *haveStatusMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected status not to be %d", m.expected)
}

// HaveStatusInRange succeeds when the status code is in [lo, hi] inclusive.
func HaveStatusInRange(lo, hi int) types.GomegaMatcher {
	return &haveStatusRangeMatcher{lo: lo, hi: hi}
}

type haveStatusRangeMatcher struct{ lo, hi int }

func (m *haveStatusRangeMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}
	return res.StatusCode >= m.lo && res.StatusCode <= m.hi, nil
}

func (m *haveStatusRangeMatcher) FailureMessage(actual interface{}) string {
	res, _ := toRecordedResponse(actual)
	return fmt.Sprintf("Expected status in [%d, %d] but got %d", m.lo, m.hi, res.StatusCode)
}

func (m *haveStatusRangeMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected status not to be in [%d, %d]", m.lo, m.hi)
}

// HaveHeader succeeds when the response contains the given header with the expected value.
func HaveHeader(key, value string) types.GomegaMatcher {
	return &haveHeaderMatcher{key: key, value: value}
}

type haveHeaderMatcher struct{ key, value string }

func (m *haveHeaderMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}
	return res.Headers.Get(m.key) == m.value, nil
}

func (m *haveHeaderMatcher) FailureMessage(actual interface{}) string {
	res, _ := toRecordedResponse(actual)
	return fmt.Sprintf("Expected header %q to be %q but got %q", m.key, m.value, res.Headers.Get(m.key))
}

func (m *haveHeaderMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected header %q not to be %q", m.key, m.value)
}

// HaveJSONBody succeeds when the response body can be JSON-decoded into expected
// and the result equals expected (using reflect.DeepEqual via json round-trip).
func HaveJSONBody(expected interface{}) types.GomegaMatcher {
	return &haveJSONBodyMatcher{expected: expected}
}

type haveJSONBodyMatcher struct{ expected interface{} }

func (m *haveJSONBodyMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}

	// Normalise both sides through JSON so type mismatches (e.g. int vs float64) are handled.
	expBytes, err := json.Marshal(m.expected)
	if err != nil {
		return false, fmt.Errorf("HaveJSONBody: cannot marshal expected: %w", err)
	}

	var expNorm, actNorm interface{}
	if err := json.Unmarshal(expBytes, &expNorm); err != nil {
		return false, err
	}
	if err := json.Unmarshal(res.BodyBytes, &actNorm); err != nil {
		return false, fmt.Errorf("HaveJSONBody: response body is not valid JSON: %w", err)
	}

	expJSON, _ := json.Marshal(expNorm)
	actJSON, _ := json.Marshal(actNorm)
	return string(expJSON) == string(actJSON), nil
}

func (m *haveJSONBodyMatcher) FailureMessage(actual interface{}) string {
	res, _ := toRecordedResponse(actual)
	return fmt.Sprintf("Expected response body to equal\n\t%+v\nbut got\n\t%s", m.expected, string(res.BodyBytes))
}

func (m *haveJSONBodyMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response body not to equal\n\t%+v", m.expected)
}

// ContainJSONKey succeeds when the response body is a JSON object containing the given key.
func ContainJSONKey(key string) types.GomegaMatcher {
	return &containJSONKeyMatcher{key: key}
}

type containJSONKeyMatcher struct{ key string }

func (m *containJSONKeyMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(res.BodyBytes, &obj); err != nil {
		return false, fmt.Errorf("ContainJSONKey: response body is not a JSON object: %w", err)
	}
	_, ok := obj[m.key]
	return ok, nil
}

func (m *containJSONKeyMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response body to contain JSON key %q", m.key)
}

func (m *containJSONKeyMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response body not to contain JSON key %q", m.key)
}

// HaveNonEmptyBody succeeds when the response body is not empty.
func HaveNonEmptyBody() types.GomegaMatcher {
	return &haveNonEmptyBodyMatcher{}
}

type haveNonEmptyBodyMatcher struct{}

func (m *haveNonEmptyBodyMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}
	return len(res.BodyBytes) > 0, nil
}

func (m *haveNonEmptyBodyMatcher) FailureMessage(_ interface{}) string {
	return "Expected response body to be non-empty"
}

func (m *haveNonEmptyBodyMatcher) NegatedFailureMessage(_ interface{}) string {
	return "Expected response body to be empty"
}

// MatchJSONSchema succeeds when every key present in the model type is also
// present in the response JSON (structural validation only — values are not compared).
func MatchJSONSchema(model interface{}) types.GomegaMatcher {
	return &matchJSONSchemaMatcher{model: model}
}

type matchJSONSchemaMatcher struct{ model interface{} }

func (m *matchJSONSchemaMatcher) Match(actual interface{}) (bool, error) {
	res, err := toRecordedResponse(actual)
	if err != nil {
		return false, err
	}

	// Build the set of keys expected by the model (top-level JSON keys).
	modelBytes, err := json.Marshal(m.model)
	if err != nil {
		return false, err
	}
	var modelMap map[string]json.RawMessage
	if err := json.Unmarshal(modelBytes, &modelMap); err != nil {
		// Model is not an object; skip structural check.
		return true, nil
	}

	var bodyMap map[string]json.RawMessage
	if err := json.Unmarshal(res.BodyBytes, &bodyMap); err != nil {
		return false, fmt.Errorf("MatchJSONSchema: response is not a JSON object: %w", err)
	}

	for k := range modelMap {
		if _, ok := bodyMap[k]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (m *matchJSONSchemaMatcher) FailureMessage(actual interface{}) string {
	res, _ := toRecordedResponse(actual)
	return fmt.Sprintf("Expected response to structurally match schema for %T\nGot: %s", m.model, string(res.BodyBytes))
}

func (m *matchJSONSchemaMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf("Expected response not to structurally match schema for %T", m.model)
}

func toRecordedResponse(actual interface{}) (*RecordedResponse, error) {
	switch v := actual.(type) {
	case *RecordedResponse:
		return v, nil
	default:
		return nil, fmt.Errorf("expected *gswag.RecordedResponse, got %T", actual)
	}
}
