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
	request.Cookies = append(request.Cookies, model.Cookie{
		SingleItemMap: model.SingleItemMap{
			Key:   util.StringPtr("best kind of cookie"),
			Value: util.StringPtr("Chocolate Chip"),
		},
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

	// Extract the key/value from the cookies.
	var cookieMaps []model.SingleItemMap
	for _, cookie := range request.Cookies {
		cookieMaps = append(cookieMaps, cookie.SingleItemMap)
	}
	if !util.Any(cookieMaps, func(key, value *string) bool {
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

func TestBodyToHeaderTransform(t *testing.T) {
	request := util.MakeRequest(t)
	response := util.MakeResponse(t)

	// TODO: disallow creating patterns that don't have zero or one capture groups
	requestTransform := BodyToHeaderTransform{
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
	requestTransform := BodyToHeaderTransform{
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

func TestHeaderPatternToHeaderTransform(t *testing.T) {
	request := util.MakeRequest(t)
	response := util.MakeResponse(t)

	// TODO: disallow creating patterns that don't have zero or one capture groups
	requestTransforms := []HeaderToHeaderTransform{
		{
			ResponseKey: "Role",
			Pattern:     "I'm in the (.+) role",
			RequestKey:  "Role",
			Before:      "user(",
			After:       "-ROLE)",
		},
		{
			ResponseKey: "AccountNumber",
			Pattern:     ".+",
			RequestKey:  "account",
			// omitting before/after
		},
		{
			// omitting ResponseKey
			Pattern:    "token-(?P<auth>[\\w-]+-\\d{5})",
			RequestKey: "Authorization-ID",
			Before:     "user(SESSION-",
			After:      ")",
		},
	}

	// When the headers don't match none of the transforms apply
	for _, transform := range requestTransforms {
		replacementTransform := transform.T(request).T(response)
		if replacementTransform != transform {
			// the transform thinks it found a match and produced a HeaderInjectionTransform
			t.Errorf("The replacementTransform wasn't the same as the original: %#v, %#v", replacementTransform, transform)
		}
	}

	// Deliberately define headers in the wrong order to find any ordering
	// constraints that could be introduced with a future regression.

	//Add a token that matches the third transform
	response.Headers = append(response.Headers, model.SingleItemMap{
		Key:   util.StringPtr("Session-27sfalkjl2k323"),
		Value: util.StringPtr("token-XMyAuthX-98765"),
	})
	// Add an accunt number that matches the second transform
	response.Headers = append(response.Headers, model.SingleItemMap{
		Key:   util.StringPtr("AccountNumber"),
		Value: util.StringPtr("some account number"),
	})
	// Add a role header to match the first transform
	response.Headers = append(response.Headers, model.SingleItemMap{
		Key:   util.StringPtr("Role"),
		Value: util.StringPtr("I'm in the accounting role"),
	})

	// let's take a look at the replacements that are produced
	headerTransforms := []HeaderInjectionTransform{}
	// When the headers match then the transforms should all swap themselves out
	// for HeaderInjectionTransforms
	for i, transform := range requestTransforms {
		replacementTransform := transform.T(request).T(response)
		if replacementTransform == transform {
			t.Fatalf("Expected a response header to match transform #%d and to replace it with a HeaderInjectionTransform replacement", i)
		}
		headerTransforms = append(headerTransforms, *replacementTransform.(*HeaderInjectionTransform))
	}

	if headerTransforms[0].Key != "Role" || headerTransforms[0].Value != "user(accounting-ROLE)" {
		t.Errorf("Wrong transform found: %#v", headerTransforms[0])
	}
	if headerTransforms[1].Key != "account" || headerTransforms[1].Value != "some account number" {
		t.Errorf("Wrong transform found: %#v", headerTransforms[1])
	}
	if headerTransforms[2].Key != "Authorization-ID" || headerTransforms[2].Value != "user(SESSION-XMyAuthX-98765)" {
		t.Errorf("Wrong transform found: %#v", headerTransforms[2])
	}

	// And now let's apply them
	for _, transform := range headerTransforms {
		transform.T(request).T(response)
	}

	// And check that the transforms did their work
	if !util.Any(request.Headers, func(key, value *string) bool {
		return *key == "Role" && *value == "user(accounting-ROLE)"
	}) {
		t.Error("no header matched the first transform")
	}
	if !util.Any(request.Headers, func(key, value *string) bool {
		return *key == "account" && *value == "some account number"
	}) {
		t.Error("no header matched the second transform")
	}
	if !util.Any(request.Headers, func(key, value *string) bool {
		return *key == "Authorization-ID" && *value == "user(SESSION-XMyAuthX-98765)"
	}) {
		t.Error("no header matched the third transform")
	}

}
