package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/raphhawk/subs/data"
)

var pageTests = []struct {
	name               string
	url                string
	expectedStatusCode int
	handler            http.HandlerFunc
	sessionData        map[string]any
	expectedHTML       string
}{
	{
		name:               "home",
		url:                "/",
		expectedStatusCode: http.StatusOK,
		handler:            testApp.HomePage,
	},
	{
		name:               "login page",
		url:                "/login",
		expectedStatusCode: http.StatusOK,
		handler:            testApp.LoginPage,
		expectedHTML:       `<h1 class="mt-5">Login</h1>`,
	},
	{
		name:               "logout",
		url:                "/logout",
		expectedStatusCode: http.StatusSeeOther,
		handler:            testApp.Logout,

		sessionData: map[string]any{
			"userId": 1,
			"user":   data.User{},
		},
	},
}

func Test_Pages(t *testing.T) {
	pathToTemplate = "./templates"

	for _, e := range pageTests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", e.url, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		if len(e.sessionData) > 0 {
			for key, value := range e.sessionData {
				testApp.Session.Put(ctx, key, value)
			}
		}

		e.handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s failed; expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if len(e.expectedHTML) > 0 {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("%s failed: expected to find %s, but did not", e.name, e.expectedHTML)
			}
		}
	}
}

func TestConfig_PostLoginPage(t *testing.T) {
	pathToTemplate = "./templates"
	postedData := url.Values{
		"email":    {"admin@example.com"},
		"password": {"abcabc12431abc1243abc1243243"},
	}

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	handler := http.HandlerFunc(testApp.PostLoginPage)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Error("Wrong code returned")
	}

	if !testApp.Session.Exists(ctx, "userId") {
		t.Error("did not find userId in session")
	}
}

func TestConfig_SubscribeToPlan(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/subscribe?id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "user", data.User{
		ID:        1,
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Active:    1,
	})

	handler := http.HandlerFunc(testApp.SubscribeToPlan)
	handler.ServeHTTP(rr, req)

	testApp.Wait.Wait()

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status code of StatusSeeOther, but got %d", rr.Code)
	}
}
