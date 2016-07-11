package transforms

import (
	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"os"
	"strings"
	"testing"
)

func makeRequest(t *testing.T) *model.Request {
	fixture := "../fixtures/browse-two-github-users.har"
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	pathToFixture := cwd + "/" + fixture
	har, err := parser.HarFrom(pathToFixture)
	if err != nil {
		t.Fatal(err)
	}
	return har.Entries[0].Request
}

func TestConstantTransformReplacesInRequestURL(t *testing.T) {
	r := makeRequest(t)

	transform := &ConstantTransform{
		Search:  "JackDanger",
		Replace: "HowzitGoing",
	}

	if !strings.Contains(r.URL, "JackDanger") {
		t.Fatalf("url didn't contain expected fixture content: %s", r.URL)
	}
	if strings.Contains(r.URL, "HowzitGoing") {
		t.Fatalf("url already contained test replacment: %s", r.URL)
	}

	responseTransform := transform.T(r)
	replacementTransform := responseTransform.T(&model.Response{})

	if strings.Contains(r.URL, "JackDanger") {
		t.Errorf("url still contains fixture content: %s", r.URL)
	}
	if !strings.Contains(r.URL, "HowzitGoing") {
		t.Errorf("url was not updated with test replacment: %s", r.URL)
	}

	if replacementTransform != nil {
		t.Error("the ResponseTransform returned non-nil, which is unexpected")
	}
}

func stringPtr(ss string) *string {
	return &ss
}
func TestConstantTransformReplacesInHeadersAndCookiesAndQueryString(t *testing.T) {
	r := makeRequest(t)
	r.Headers = append(r.Headers, model.SingleItemMap{
		Key:   stringPtr("PreviousKey"),
		Value: stringPtr("PreviousValue"),
	})
	r.Cookies = append(r.Cookies, model.SingleItemMap{
		Key:   stringPtr("best kind of cookie"),
		Value: stringPtr("Chocolate Chip"),
	})
	r.QueryString = append(r.QueryString, model.SingleItemMap{
		Key:   stringPtr("timezone"),
		Value: stringPtr("_TIMEZONE_"),
	})

	transforms := []ConstantTransform{
		{
			Search:  "PreviousKey",
			Replace: "usedtobePreviousKey",
		},
		{
			Search:  "PreviousValue",
			Replace: "nope.gif",
		},
		{
			Search:  "Chocolate Chip",
			Replace: "Peanut Butter",
		},
		{
			Search:  "Chocolate Chip",
			Replace: "Peanut Butter",
		},
		{
			Search:  "_TIMEZONE_",
			Replace: "America/Los_Angeles",
		},
	}

	for _, transform := range transforms {
		responseTransform := transform.T(r)
		replacementTransform := responseTransform.T(&model.Response{})
		if replacementTransform != nil {
			t.Error("the ResponseTransform returned non-nil, which is unexpected")
		}
	}

	if !any(r.Headers, func(key, value *string) bool {
		return strings.Contains(*key, "usedtobePreviousKey")
	}) {
		t.Errorf("header key is unchanged: %v", r.Headers)
	}

	if !any(r.Headers, func(key, value *string) bool {
		return strings.Contains(*value, "nope.gif")
	}) {
		t.Errorf("header value is unchanged: %v", r.Headers)
	}

	if !any(r.Cookies, func(key, value *string) bool {
		return strings.Contains(*value, "Peanut Butter")
	}) {
		t.Errorf("cookie is unchanged: %v", r.Cookies)
	}

	if !any(r.QueryString, func(key, value *string) bool {
		return *key == "timezone" && strings.Contains(*value, "America/Los_Angeles")
	}) {
		t.Errorf("querystring is unchanged: %v", r.QueryString)
	}
}

func TestHeaderInjectionTransform(t *testing.T) {
	r := makeRequest(t)

	if any(r.Headers, func(key, value *string) bool {
		return *key == "newKey" || *value == "newValue"
	}) {
		t.Errorf("New header already exists in request: %v", r.Headers)
	}

	transform := &HeaderInjectionTransform{
		Key:   "newKey",
		Value: "newValue",
	}

	transform.T(r)

	if !any(r.Headers, func(key, value *string) bool {
		return *key == "newKey" && *value == "newValue"
	}) {
		t.Errorf("New header was not added to request: %#v", r.Headers)
	}
}

type pairwiseFunc func(key, val *string) bool

func any(pairs []model.SingleItemMap, f pairwiseFunc) bool {
	for _, pair := range pairs {
		if f(pair.Key, pair.Value) {
			return true
		}
	}
	return false
}
