package transforms

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JackDanger/traffic/model"
	util "github.com/JackDanger/traffic/test"
)

func TestConstantTransformReplacesInRequestURL(t *testing.T) {
	r := util.MakeRequest(t)

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
	request := util.MakeRequest(t)
	request.Headers = append(request.Headers, model.SingleItemMap{
		Key:   stringPtr("PreviousKey"),
		Value: stringPtr("PreviousValue"),
	})
	request.Cookies = append(request.Cookies, model.SingleItemMap{
		Key:   stringPtr("best kind of cookie"),
		Value: stringPtr("Chocolate Chip"),
	})
	request.QueryString = append(request.QueryString, model.SingleItemMap{
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
		responseTransform := transform.T(request)
		replacementTransform := responseTransform.T(&model.Response{})
		if replacementTransform != nil {
			t.Error("the ResponseTransform returned non-nil, which is unexpected")
		}
	}

	if !any(request.Headers, func(key, value *string) bool {
		return strings.Contains(*key, "usedtobePreviousKey")
	}) {
		t.Errorf("header key is unchanged: %v", request.Headers)
	}

	if !any(request.Headers, func(key, value *string) bool {
		return strings.Contains(*value, "nope.gif")
	}) {
		t.Errorf("header value is unchanged: %v", request.Headers)
	}

	if !any(request.Cookies, func(key, value *string) bool {
		return strings.Contains(*value, "Peanut Butter")
	}) {
		t.Errorf("cookie is unchanged: %v", request.Cookies)
	}

	if !any(request.QueryString, func(key, value *string) bool {
		return *key == "timezone" && strings.Contains(*value, "America/Los_Angeles")
	}) {
		t.Errorf("querystring is unchanged: %v", request.QueryString)
	}
}

func TestHeaderInjectionTransform(t *testing.T) {
	request := util.MakeRequest(t)

	if any(request.Headers, func(key, value *string) bool {
		return *key == "newKey" || *value == "newValue"
	}) {
		t.Errorf("New header already exists in request: %v", request.Headers)
	}

	transform := &HeaderInjectionTransform{
		Key:   "newKey",
		Value: "newValue",
	}

	transform.T(request)

	if !any(request.Headers, func(key, value *string) bool {
		return *key == "newKey" && *value == "newValue"
	}) {
		t.Errorf("New header was not added to request: %#v", request.Headers)
	}
}

func TestResponseBodyToRequestHeaderTransform(t *testing.T) {
	request := util.MakeRequest(t)
	response := util.MakeResponse(t)

	// TODO: disallow creating patterns that don't have exactly one capture group
	requestTransform := &ResponseBodyToRequestHeaderTransform{
		Pattern:    "token-(?P<auth>[\\w-]+-\\d{5})",
		HeaderName: "Authorization-ID",
	}

	// When the response body does not match
	responseTransform := requestTransform.T(request)
	replacementTransform := responseTransform.T(response)

	if replacementTransform != nil {
		t.Error("The replacementTransform wasn't nil which means the transform thinks it found a match and produced a HeaderInjectionTransform")
	}

	response.ContentBody = stringPtr(`{
		"user": {
			"name": "Sarah",
			"role": "CTO"
		},
		"session": {
			"token": "token-ABC123-00123"
		}
	}`)

	// When the response body matches
	responseTransform = requestTransform.T(request)
	replacementTransform = responseTransform.T(response)
	if replacementTransform == nil {
		t.Error("Expected the response body to match and to return a HeaderInjectionTransform replacement")
	}

	replacement, ok := replacementTransform.(HeaderInjectionTransform)
	if !ok {
		t.Errorf("expected replacementTransform to be a HeaderInjectionTransform, was: %#v", reflect.TypeOf(replacementTransform))
	}

	if replacementTransform == nil {
		t.Error("The replacementTransform should have been a HeaderInjectionTransform")
	}

	if replacement.Key != "Authorization-ID" {
		t.Errorf("Expected HeaderName to be Authorization-ID, got: %s", replacement.Key)
	}
	if replacement.Value != "ABC123-00123" {
		t.Errorf("Expected HeaderName to be ABC123-00123, got: %s", replacement.Value)
	}

}

// test helpers

type pairwiseFunc func(key, val *string) bool

func any(pairs []model.SingleItemMap, f pairwiseFunc) bool {
	for _, pair := range pairs {
		if f(pair.Key, pair.Value) {
			return true
		}
	}
	return false
}
