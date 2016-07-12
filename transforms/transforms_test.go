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

	if replacementTransform != transform {
		t.Error("the ResponseTransform returned something other than the original RequestTransform, which is unexpected")
	}
}

func TestConstantTransformReplacesInHeadersAndCookiesAndQueryString(t *testing.T) {
	request := util.MakeRequest(t)
	request.Headers = append(request.Headers, model.SingleItemMap{
		Key:   util.StringPtr("PreviousKey"),
		Value: util.StringPtr("PreviousValue"),
	})
	request.Cookies = append(request.Cookies, model.SingleItemMap{
		Key:   util.StringPtr("best kind of cookie"),
		Value: util.StringPtr("Chocolate Chip"),
	})
	request.QueryString = append(request.QueryString, model.SingleItemMap{
		Key:   util.StringPtr("timezone"),
		Value: util.StringPtr("_TIMEZONE_"),
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
		if replacementTransform != &transform {
			t.Error("the ResponseTransform returned something other than the original RequestTransform, which is unexpected")
		}
	}

	if !util.Any(request.Headers, func(key, value *string) bool {
		return strings.Contains(*key, "usedtobePreviousKey")
	}) {
		t.Errorf("header key is unchanged: %v", request.Headers)
	}

	if !util.Any(request.Headers, func(key, value *string) bool {
		return strings.Contains(*value, "nope.gif")
	}) {
		t.Errorf("header value is unchanged: %v", request.Headers)
	}

	if !util.Any(request.Cookies, func(key, value *string) bool {
		return strings.Contains(*value, "Peanut Butter")
	}) {
		t.Errorf("cookie is unchanged: %v", request.Cookies)
	}

	if !util.Any(request.QueryString, func(key, value *string) bool {
		return *key == "timezone" && strings.Contains(*value, "America/Los_Angeles")
	}) {
		t.Errorf("querystring is unchanged: %v", request.QueryString)
	}
}

func TestHeaderInjectionTransform(t *testing.T) {
	request := util.MakeRequest(t)

	if util.Any(request.Headers, func(key, value *string) bool {
		return *key == "newKey" || *value == "newValue"
	}) {
		t.Errorf("New header already exists in request: %v", request.Headers)
	}

	transform := &HeaderInjectionTransform{
		Key:   "newKey",
		Value: "newValue",
	}

	transform.T(request)

	if !util.Any(request.Headers, func(key, value *string) bool {
		return *key == "newKey" && *value == "newValue"
	}) {
		t.Errorf("New header was not added to request: %#v", request.Headers)
	}
}

func TestResponseBodyToRequestHeaderTransform(t *testing.T) {
	request := util.MakeRequest(t)
	response := util.MakeResponse(t)

	// TODO: disallow creating patterns that don't have zero or one capture groups
	requestTransform := ResponseBodyToRequestHeaderTransform{
		Pattern:    "token-(?P<auth>[\\w-]+-\\d{5})",
		HeaderName: "Authorization-ID",
	}

	// When the response body does not match
	responseTransform := requestTransform.T(request)
	replacementTransform := responseTransform.T(response)

	if replacementTransform != requestTransform {
		// the transform thinks it found a match and produced a HeaderInjectionTransform
		t.Errorf("The replacementTransform wasn't the same as the original: %#v, %#v", replacementTransform, requestTransform)
	}

	response.ContentBody = util.StringPtr(`{
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

	if replacementTransform == requestTransform {
		t.Error("The replacementTransform should have been a new transform, not the original")
	}

	if replacement.Key != "Authorization-ID" {
		t.Errorf("Expected HeaderName to be Authorization-ID, got: %s", replacement.Key)
	}
	if replacement.Value != "ABC123-00123" {
		t.Errorf("Expected HeaderName to be ABC123-00123, got: %s", replacement.Value)
	}

}
func TestResponseWithComplexBodyToRequestHeaderTransform(t *testing.T) {
	request := util.MakeRequest(t)
	response := util.MakeResponse(t)

	// TODO: disallow creating patterns that don't have zero or one capture groups
	requestTransform := ResponseBodyToRequestHeaderTransform{
		Pattern:    "token-(?P<auth>[\\w-]+-\\d{5})",
		HeaderName: "Authorization-ID",
		Before:     "user(OWNER-",
		After:      ")",
	}

	// When the response body does not match
	responseTransform := requestTransform.T(request)
	replacementTransform := responseTransform.T(response)

	if replacementTransform != requestTransform {
		// the transform thinks it found a match and produced a HeaderInjectionTransform
		t.Errorf("The replacementTransform wasn't the same as the original: %#v, %#v", replacementTransform, requestTransform)
	}

	response.ContentBody = util.StringPtr(`{
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

	if replacementTransform == requestTransform {
		t.Error("The replacementTransform should have been a new transform, not the original")
	}

	if replacement.Key != "Authorization-ID" {
		t.Errorf("Expected HeaderName to be Authorization-ID, got: %s", replacement.Key)
	}
	if replacement.Value != "user(OWNER-ABC123-00123)" {
		t.Errorf("Expected HeaderName to be user(OWNER-ABC123-00123), got: %s", replacement.Value)
	}

}
