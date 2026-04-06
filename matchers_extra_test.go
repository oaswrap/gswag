package gswag

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func respWith(body string, status int, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

func TestHaveStatus_FailureMessages(t *testing.T) {
	resp := respWith("oops", 500, nil)
	m := HaveStatus(200)
	ok, err := m.Match(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected not ok")
	}
	fm := m.(*haveStatusMatcher).FailureMessage(resp)
	if !strings.Contains(fm, "Expected status 200 but got 500") {
		t.Fatalf("unexpected failure message: %s", fm)
	}
}

func TestHaveStatusInRange_Failure(t *testing.T) {
	resp := respWith("", 302, nil)
	m := HaveStatusInRange(200, 299)
	ok, _ := m.Match(resp)
	if ok {
		t.Fatalf("expected not ok")
	}
	fm := m.(*haveStatusRangeMatcher).FailureMessage(resp)
	if !strings.Contains(fm, "Expected status in [200, 299] but got 302") {
		t.Fatalf("unexpected msg: %s", fm)
	}
}

func TestHaveHeaderAndJSONMatchers_FailureMessages(t *testing.T) {
	resp := respWith(`notjson`, 200, nil)

	// header missing
	hm := HaveHeader("X-Req", "v")
	ok, _ := hm.Match(resp)
	if ok {
		t.Fatalf("expected header match to fail")
	}
	hfm := hm.(*haveHeaderMatcher).FailureMessage(resp)
	if !strings.Contains(hfm, "Expected header") {
		t.Fatalf("unexpected header failure: %s", hfm)
	}

	// JSON body equality failure message
	jb := HaveJSONBody(map[string]interface{}{"a": 1})
	ok2, err := jb.Match(resp)
	if err == nil && ok2 {
		t.Fatalf("expected JSON match to fail")
	}
	jfm := jb.(*haveJSONBodyMatcher).FailureMessage(resp)
	if !strings.Contains(jfm, "Expected response body to equal") {
		t.Fatalf("unexpected json failure message: %s", jfm)
	}

	// ContainJSONKey failure message
	ck := ContainJSONKey("x")
	ok3, _ := ck.Match(resp)
	if ok3 {
		t.Fatalf("expected contain json key to fail")
	}
	if !strings.Contains(ck.(*containJSONKeyMatcher).FailureMessage(nil), "contain JSON key") {
		t.Fatalf("unexpected contain json key failure")
	}
}

func TestHaveNonEmptyBodyAndMatchJSONSchemaFailures(t *testing.T) {
	empty := respWith("", 200, nil)
	nb := HaveNonEmptyBody()
	ok, _ := nb.Match(empty)
	if ok {
		t.Fatalf("expected empty body to be false")
	}
	if !strings.Contains(nb.(*haveNonEmptyBodyMatcher).FailureMessage(nil), "non-empty") {
		t.Fatalf("unexpected non-empty failure")
	}

	// MatchJSONSchema: model has field 'id' but body lacks it
	body := respWith(`{"name":"x"}`, 200, nil)
	ms := MatchJSONSchema(struct {
		ID string `json:"id"`
	}{})
	ok2, err := ms.Match(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok2 {
		t.Fatalf("expected schema match to fail")
	}
	if !strings.Contains(ms.(*matchJSONSchemaMatcher).FailureMessage(body), "structurally match schema") {
		t.Fatalf("unexpected schema failure message")
	}
}

func TestNegatedFailureMessages(t *testing.T) {
	jb := HaveJSONBody(map[string]interface{}{"a": 1})
	if !strings.Contains(jb.(*haveJSONBodyMatcher).NegatedFailureMessage(nil), "not to equal") {
		t.Fatalf("unexpected negated message for HaveJSONBody")
	}

	nb := HaveNonEmptyBody()
	if !strings.Contains(nb.(*haveNonEmptyBodyMatcher).NegatedFailureMessage(nil), "empty") {
		t.Fatalf("unexpected negated message for HaveNonEmptyBody")
	}
}
