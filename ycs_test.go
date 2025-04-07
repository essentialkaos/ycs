package ycs

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const TEST_PORT = "56123"

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type YCSSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&YCSSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *YCSSuite) SetUpSuite(c *C) {
	apiURL = "http://127.0.0.1:" + TEST_PORT

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":" + TEST_PORT, Handler: mux}

	mux.HandleFunc("GET /services", handlerServices)
	mux.HandleFunc("GET /incidents", handlerIncidents)
	mux.HandleFunc("GET /incidents/972", handlerIncident)

	go server.ListenAndServe()

	time.Sleep(time.Second)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *YCSSuite) TestInit(c *C) {
	SetUserAgent("Test", "1.2.3")
	SetLimit(100)
	SetRequestTimeout(0.1)

	c.Assert(engine, NotNil)

	engine = nil
}

func (s *YCSSuite) TestGetServices(c *C) {
	services, err := GetServices(LANG_RU)

	c.Assert(err, IsNil)
	c.Assert(services, HasLen, 104)

	c.Assert(services.InRegion(REGION_RU), HasLen, 74)
	c.Assert(services.IDs(), HasLen, 104)
	c.Assert(services.Names(), HasLen, 104)
}

func (s *YCSSuite) TestGetIncidents(c *C) {
	incidents, err := GetIncidents(IncidentsRequest{
		Lang:   LANG_RU,
		From:   time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local),
		To:     time.Date(2025, 01, 8, 0, 0, 0, 0, time.Local),
		Status: STATUS_OPEN,
		Region: "all",
		Zones:  []string{ZONE_RU_A},
	})

	c.Assert(err, IsNil)
	c.Assert(incidents, HasLen, 20)
	c.Assert(incidents.HasOpen(), Equals, true)

	incidents = Incidents{}
	c.Assert(incidents.HasOpen(), Equals, false)
}

func (s *YCSSuite) TestGetIncident(c *C) {
	incident, err := GetIncident(972, LANG_EN)

	c.Assert(err, IsNil)
	c.Assert(incident, NotNil)

	c.Assert(incident.IsResolved(), Equals, true)
	c.Assert(incident.Duration().String(), Equals, "7h2m0s")
	c.Assert(incident.URL(LANG_EN), Equals, "https://status.yandex.cloud/en/incidents/972")
	c.Assert(incident.ReportMarkdown(), Not(Equals), "")
	c.Assert(incident.RegionList(), DeepEquals, []string{"ru"})
	c.Assert(incident.ZoneList(), DeepEquals, []string{"ru-central1-a"})
	c.Assert(incident.ServiceList(), DeepEquals, []string{"Compute Cloud", "Virtual Private Cloud", "Network Load Balancer", "Managed Service for Kubernetes®", "Monitoring", "Managed Service for PostgreSQL", "Managed Service for ClickHouse®", "Managed Service for MongoDB", "Managed Service for Valkey™", "Data Processing", "SpeechKit", "Translate", "Vision OCR", "Managed Service for YDB", "Cloud Interconnect", "Data Transfer", "DataSphere", "Managed Service for Apache Kafka®", "Managed Service for Elasticsearch", "Application Load Balancer", "Cloud DNS", "Cloud CDN", "Cloud Logging", "Managed Service for Greenplum®", "Data Streams", "Managed Service for GitLab", "Cloud Desktop", "Yandex Query", "Managed Service for OpenSearch", "YandexGPT API", "Yandex Cloud Billing", "Yandex WebSQL", "Yandex Managed Service for Apache Airflow™", "Managed Service for Prometheus®", "SpeechSense", "Yandex MetaData Hub", "Foundation Models"})
	c.Assert(incident.Comments.Get(0), NotNil)
	c.Assert(incident.Comments.Get(0).Markdown(), Not(Equals), "")
	c.Assert(incident.Comments.Get(5), IsNil)

	incident = nil
	var comments Comments

	c.Assert(incident.IsResolved(), Equals, false)
	c.Assert(incident.Duration().String(), Equals, "0s")
	c.Assert(incident.URL(LANG_EN), Equals, "")
	c.Assert(incident.ReportMarkdown(), Equals, "")
	c.Assert(incident.RegionList(), IsNil)
	c.Assert(incident.ZoneList(), IsNil)
	c.Assert(incident.ServiceList(), IsNil)
	c.Assert(comments.Get(0), IsNil)
	c.Assert(comments.Get(0).Markdown(), Equals, "")
}

func (s *YCSSuite) TestErrors(c *C) {
	SetUserAgent("http-error", "1")

	_, err := GetServices(LANG_RU)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Can't get services status: API returned non-ok status code 503")

	_, err = GetIncidents(IncidentsRequest{})
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Can't get incidents: API returned non-ok status code 503")

	_, err = GetIncident(972, LANG_EN)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Can't get incident 972: API returned non-ok status code 503")

	SetUserAgent("data-error", "1")

	_, err = GetIncident(972, LANG_EN)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Can't get incident 972: Can't decode API response: invalid character 'F' looking for beginning of value")

	SetUserAgent("", "")

	apiURL = "http://127.0.0.1:9999"

	_, err = GetIncident(972, LANG_EN)
	c.Assert(err, NotNil)

	apiURL = "http://127.0.0.1:" + TEST_PORT
}

func (s *YCSSuite) TestDateParsing(c *C) {
	d := &Date{}

	err := d.UnmarshalJSON([]byte(`"2025-01-22T21:52:41Z"`))

	c.Assert(err, IsNil)
	c.Assert(d.Unix(), Equals, int64(1737582761))

	err = d.UnmarshalJSON([]byte(`null`))

	c.Assert(err, IsNil)
	c.Assert(d.IsZero(), Equals, true)

	err = d.UnmarshalJSON([]byte(`ABCD`))
	c.Assert(err, NotNil)
}

func (s *YCSSuite) TestHtml2Markdown(c *C) {
	c.Assert(htmlToMarkdown("<i>italic</i>"), Equals, "*italic*")
	c.Assert(htmlToMarkdown("<b>bold</b>"), Equals, "**bold**")
	c.Assert(htmlToMarkdown("<strong>bold</strong>"), Equals, "**bold**")
	c.Assert(htmlToMarkdown("<pre>code</pre>"), Equals, "`code`")
	c.Assert(htmlToMarkdown("<code>code-block</code>"), Equals, "```code-block```")
	c.Assert(htmlToMarkdown("<br/>"), Equals, "\n")
	c.Assert(htmlToMarkdown("<ul><li>Test</li></ul>"), Equals, "• Test")
	c.Assert(htmlToMarkdown("<ol><li>Test</li></ol>"), Equals, "1. Test")
	c.Assert(htmlToMarkdown(`<a href="https://domain.com" title="title">Link</a>`), Equals, "[Link](https://domain.com)")
	c.Assert(htmlToMarkdown(`<img src="https://domain.com/image.png" />`), Equals, "![IMG](https://domain.com/image.png)")
}

// ////////////////////////////////////////////////////////////////////////////////// //

func handlerServices(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	data, _ := os.ReadFile("testdata/services.json")
	rw.Write(data)
}

func handlerIncidents(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	data, _ := os.ReadFile("testdata/incidents.json")
	rw.Write(data)
}

func handlerIncident(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	data, _ := os.ReadFile("testdata/incident.json")
	rw.Write(data)
}

func writeErrorResponse(rw http.ResponseWriter, r *http.Request) bool {
	if strings.Contains(r.Header.Get("User-Agent"), "http-error") {
		rw.WriteHeader(503)
		return true
	}

	if strings.Contains(r.Header.Get("User-Agent"), "data-error") {
		rw.WriteHeader(200)
		rw.Write([]byte(`FFFF`))
		return true
	}

	return false
}
