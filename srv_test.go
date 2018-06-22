package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testQuery struct {
	// request doby
	request string
	// expecded status code
	code int
	// expected data. It should not be a pointer for sake of simplify
	data interface{}
}

func TestConfigServer(t *testing.T) {
	queries := []testQuery{
		{
			// invalid json
			request: "^_^",
			code:    http.StatusBadRequest,
		},
		{
			request: `"^_^"`,
			code:    http.StatusBadRequest,
		},
		{
			request: `{"Type": "does not exist", "Data": "whatever"}`,
			code:    http.StatusNotFound,
		},
		{
			request: `{"Type": "Develop.mr_robot", "Data": "Database.processing"}`,
			code:    http.StatusOK,
			// @TODO
			data: "",
		},
		{
			request: `{"Type": "Test.vpn", "Data": "Rabbit.log"}`,
			code:    http.StatusOK,
			// @TODO
			data: "wat?",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(configHandler))
	defer ts.Close()

	for _, query := range queries {
		checkQuery(t, ts, query)
	}
}

func checkQuery(t *testing.T, ts *httptest.Server, query testQuery) {
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(query.request))
	if err != nil {
		t.Fatalf("request '%v': failed to perform http request: %v", query.request, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != query.code {
		t.Errorf("request '%v': unexpected status code %v(%v expected)", query.request, resp.StatusCode, query.code)
		return
	}

	if query.code != http.StatusOK {
		t.Logf("request '%v': passed", query.request)
		return
	}

	if query.data == nil {
		t.Fatalf("request '%v': no reply data defined for status 200", query.request)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("request '%v': failed to read body: %v", query.request, err)
	}

	// Indents and order of fields should not change anything.
	// We will unmarshal reply into a value with the expected type and compare the result with reference.
	data := reflect.New(reflect.ValueOf(query.data).Type())
	err = json.Unmarshal(body, data.Interface())
	if err != nil {
		t.Errorf("request '%v': failed to unmarshal reply: %v\nraw body:\n`%v`", query.request, err, string(body))
		return
	}

	if !reflect.DeepEqual(query.data, reflect.Indirect(data).Interface()) {
		bytes, _ := json.MarshalIndent(query.data, "", "  ")
		t.Errorf("request '%v': reply does not match expectations, raw reply:\n`%v`\nbut\n`%v`\nexpected",
			query.request, string(body), string(bytes))
		return
	}

	t.Logf("request '%v': passed", query.request)
}
