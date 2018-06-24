package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	dbConfig := os.Getenv("TEST_CONFIG_DB")
	if dbConfig == "" {
		log.Fatal("please, set TEST_CONFIG_DB variable before running the tests")
	}

	var err error
	db, err = gorm.Open("postgres", dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to the test database: %v", err)
	}

	flag.Parse()
	if !testing.Verbose() {
		db.LogMode(false)
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type testQuery struct {
	// request doby
	request string
	// expecded status code
	code int
	// expected json reply for 200 status
	data string
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
			request: `{"Type": "", "Data": "whatever"}`,
			code:    http.StatusBadRequest,
		},
		{
			request: `{"Type": "does not exist", "Data": "whatever"}`,
			code:    http.StatusNotFound,
		},
		{
			request: `{"Type": "database.postgres", "Data": "service.test"}`,
			code:    http.StatusOK,
			data: `
			{
				"host": "localhost",
				"port": "5432",
				"database": "devdb",
				"user": "mr_robot",
				"password": "secret",
				"schema": "public"
			}`,
		},
		{
			request: `{"Type": "rabbit.log", "Data": "service.test"}`,
			code:    http.StatusOK,
			// swapped order of fields
			data: `
			{
				"user": "guest",
				"password": "guest",
				"host": "10.0.5.42",
				"port": "5671",
				"virtualhost": "/"
			}`,
		},
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", newConfigServer(db).handle)
	ts := httptest.NewServer(r)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("request '%v': failed to read body: %v", query.request, err)
	}

	// Indents and order of fields should not change anything.
	// We will unmarshal both the expected data and the reply for comparison.
	var reply, expected interface{}
	err = json.Unmarshal(body, &reply)
	if err != nil {
		t.Errorf("request '%v': failed to unmarshal reply: %v\nraw body:\n`%v`", query.request, err, string(body))
		return
	}

	err = json.Unmarshal([]byte(query.data), &expected)
	if err != nil {
		t.Errorf("request '%v': failed to unmarshal expected data: %v", query.request, err)
		return
	}

	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("request '%v': reply does not match expectations, raw reply:\n`%v`\nbut\n`%v`\nexpected",
			query.request, string(body), query.data)
		return
	}

	t.Logf("request '%v': passed", query.request)
}
