package ycs

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/req"
	"github.com/essentialkaos/ek/v13/reutil"
	"github.com/essentialkaos/ek/v13/sliceutil"
	"github.com/essentialkaos/ek/v13/strutil"
	"github.com/essentialkaos/ek/v13/timeutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// UA is HTTP client user-agent
const UA = "EK|YCS.go"

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	LANG_RU = "ru"
	LANG_EN = "en"
)

const (
	REGION_ALL = "all"
	REGION_RU  = "ru"
	REGION_KZ  = "kz"
)

const (
	ZONE_RU_A = "ru-central1-a"
	ZONE_RU_B = "ru-central1-b"
	ZONE_RU_C = "ru-central1-c"
	ZONE_RU_D = "ru-central1-d"
	ZONE_KZ_A = "kz1-a"
)

const (
	STATUS_OPEN        = "open"
	STATUS_RESOLVED    = "resolved"
	STATUS_WITH_REPORT = "withReport"
)

const (
	TYPE_INVESTIGATION = "investigation"
	TYPE_UPDATE        = "update"
	TYPE_RESOLVED      = "resolved"
)

const (
	LEVEL_MINOR       = "Minor"       // 1
	LEVEL_UNAVAILABLE = "Unavailable" // 2
)

const (
	LEVEL_ID_MINOR       uint8 = 1
	LEVEL_ID_UNAVAILABLE uint8 = 2
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Date is JSON date
type Date struct {
	time.Time
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Service contains info about service
type Service struct {
	Name             string    `json:"name"`
	FullName         string    `json:"fullName"`
	Slug             string    `json:"slug"`
	Description      string    `json:"description"`
	IAMFlag          string    `json:"iamFlag,omitempty"`
	Status           string    `json:"status"`
	DocURL           string    `json:"docUrl"`
	PricesURL        string    `json:"pricesUrl"`
	ConsoleURL       string    `json:"consoleUrl"`
	InstallationCode string    `json:"installationCode"`
	Icon             string    `json:"icon"`
	IconName         string    `json:"iconName"`
	OrderNumber      uint      `json:"orderNumber"`
	CategoryID       uint      `json:"categoryId"`
	ID               uint      `json:"id"`
	PageID           uint      `json:"pageId"`
	CreatedAt        Date      `json:"createdAt"`
	UpdatedAt        Date      `json:"updatedAt"`
	IsProduct        bool      `json:"isProduct"`
	Incidents        Incidents `json:"incidents"`
}

// Services is slice with services
type Services []*Service

// Incident contains info about incident
type Incident struct {
	ID                  uint     `json:"id"`
	Title               string   `json:"title"`
	Report              string   `json:"report,omitempty"`
	Status              string   `json:"status"`
	IsReportPublished   bool     `json:"isReportPublished,omitempty"`
	LevelID             uint8    `json:"levelId"`
	StartDate           Date     `json:"startDate"`
	EndDate             Date     `json:"endDate"`
	CreatedAt           Date     `json:"createdAt"`
	UpdatedAt           Date     `json:"updatedAt"`
	ReportPublishedTime Date     `json:"reportPublishedTime"`
	Level               *Level   `json:"level"`
	Zones               Zones    `json:"zones"`
	Regions             Regions  `json:"installations"`
	Services            Services `json:"services"`
	Comments            Comments `json:"comments"`
}

// Incidents is a slice with incidents
type Incidents []*Incident

// Level contains info about incident level
type Level struct {
	Level     uint8  `json:"level"`
	Label     string `json:"label"`
	Theme     string `json:"theme"`
	CreatedAt Date   `json:"createdAt"`
	UpdatedAt Date   `json:"updatedAt"`
}

// Zone contains info about zone (RU/KZ)
type Zone struct {
	InstallationID uint    `json:"installationId"`
	ID             string  `json:"id"`
	CreatedAt      Date    `json:"createdAt"`
	UpdatedAt      Date    `json:"updatedAt"`
	Region         *Region `json:"installation"`
}

// Zones is slice with zones
type Zones []*Zone

// Region contains info about installation region
type Region struct {
	Code  string `json:"code"`
	Zones Zones  `json:"zones"`
}

// Regions is a slice with installation regions
type Regions []*Region

type Comment struct {
	ID         uint   `json:"id"`
	IncidentID uint   `json:"incidentId"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	CreatedAt  Date   `json:"createdAt"`
	UpdatedAt  Date   `json:"updatedAt"`
}

// Comments is a slice with comments
type Comments []*Comment

// ////////////////////////////////////////////////////////////////////////////////// //

// IncidentsRequest contains incident request info
type IncidentsRequest struct {
	Lang   string
	From   time.Time
	To     time.Time
	Status string
	Region string
	Zones  []string
}

// ////////////////////////////////////////////////////////////////////////////////// //

// AllLangs is a slice with all supported languages
var AllLangs = []string{LANG_EN, LANG_RU}

// AllRegions is a slice with all regions
var AllRegions = []string{REGION_ALL, REGION_KZ, REGION_RU}

// AllZones is a slice with all availability zones
var AllZones = []string{ZONE_KZ_A, ZONE_RU_A, ZONE_RU_B, ZONE_RU_C, ZONE_RU_D}

// ////////////////////////////////////////////////////////////////////////////////// //

// engine is HTTP client
var engine *req.Engine

// apiURL is Yandex.Cloud status API URL
var apiURL = "https://status.yandex.cloud/api"

var (
	htmlTagStartRegex = regexp.MustCompile(`<(strong|pre|code|ol|ul|li|br|i|b|p)[^>]*\/?>($|\n)?`)
	htmlTagEndRegex   = regexp.MustCompile(`<\/(strong|pre|code|ol|ul|li|i|b|p)\/?>`)
	htmlImgTagRegex   = regexp.MustCompile(`<img src=\"([^"]+)\"[^>]+\>`)
	htmlLinkTagRegex  = regexp.MustCompile(`<a href=\"([^"]+)\"[^>]*>([^<]+)<\/a>`)
)

// ////////////////////////////////////////////////////////////////////////////////// //

// SetUserAgent sets user agent
func SetUserAgent(app, version string) {
	initEngine()
	engine.SetUserAgent(app, version, UA+"/1")
}

// SetLimit sets a hard limit on the number of requests per second
func SetLimit(rps float64) {
	initEngine()
	engine.SetLimit(rps)
}

// SetRequestTimeout sets request timeout
func SetRequestTimeout(timeout float64) {
	initEngine()
	engine.SetRequestTimeout(timeout)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetServices returns status of all services
func GetServices(lang string) (Services, error) {
	initEngine()

	resp := Services{}

	err := sendRequest(
		"/services",
		req.Query{
			"incidents": "all",
			"lang":      strutil.Q(lang, LANG_RU),
		},
		&resp,
	)

	if err != nil {
		return nil, fmt.Errorf("Can't get services status: %w", err)
	}

	return resp, nil
}

// GetIncidents returns slice with incidents
func GetIncidents(req IncidentsRequest) (Incidents, error) {
	initEngine()

	resp := &struct {
		Items Incidents `json:"items"`
	}{}

	err := sendRequest(
		"/incidents",
		convertIncidentsRequest(req),
		&resp,
	)

	if err != nil {
		return nil, fmt.Errorf("Can't get incidents: %w", err)
	}

	return resp.Items, nil
}

// GetIncidents returns slice with incidents
func GetIncident(id uint, lang string) (*Incident, error) {
	initEngine()

	resp := &Incident{}

	err := sendRequest(
		fmt.Sprintf("/incidents/%d", id),
		req.Query{"lang": strutil.Q(lang, LANG_RU)},
		&resp,
	)

	if err != nil {
		return nil, fmt.Errorf("Can't get incident %d: %w", id, err)
	}

	return resp, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// HasOpen returns true if slice contains open incident
func (i Incidents) HasOpen() bool {
	for _, ii := range i {
		if ii.Status == STATUS_OPEN {
			return true
		}
	}

	return false
}

// IsResolved returns true if incident is resolved
func (i *Incident) IsResolved() bool {
	return i != nil && i.Status == STATUS_RESOLVED
}

// Duration returns incident duration
func (i *Incident) Duration() time.Duration {
	if i == nil || i.EndDate.IsZero() {
		return 0
	}

	return i.EndDate.Sub(i.StartDate.Time)
}

// URL returns URL of incident page
func (i *Incident) URL(lang string) string {
	if i == nil || i.ID == 0 {
		return ""
	}

	return fmt.Sprintf(
		"https://status.yandex.cloud/%s/incidents/%d",
		lang, i.ID,
	)
}

// ReportMarkdown converts report HTML to Markdown
func (i *Incident) ReportMarkdown() string {
	if i == nil {
		return ""
	}

	return htmlToMarkdown(i.Report)
}

// RegionList returns slice with all regions affected by the incident
func (i *Incident) RegionList() []string {
	if i == nil || len(i.Regions) == 0 {
		return nil
	}

	var result []string

	for _, r := range i.Regions {
		result = append(result, r.Code)
	}

	return result
}

// ZoneList returns slice with all zones affected by the incident
func (i *Incident) ZoneList() []string {
	if i == nil || len(i.Regions) == 0 {
		return nil
	}

	var result []string

	for _, r := range i.Regions {
		for _, z := range r.Zones {
			result = append(result, z.ID)
		}
	}

	return result
}

// ServiceList returns slice with all services affected by the incident
func (i *Incident) ServiceList() []string {
	if i == nil || len(i.Services) == 0 {
		return nil
	}

	var result []string

	for _, s := range i.Services {
		result = append(result, s.Name)
	}

	return result
}

// InRegion filters services and returns only services in a given region (installation)
func (s Services) InRegion(code string) Services {
	return sliceutil.Filter(s, func(ss *Service, _ int) bool {
		return ss.InstallationCode == code
	})
}

// IDs returns slice with IDs of services
func (s Services) IDs() []uint {
	var result []uint

	for _, ss := range s {
		result = append(result, ss.ID)
	}

	return result
}

// Names returns slice with names of services
func (s Services) Names() []string {
	var result []string

	for _, ss := range s {
		result = append(result, ss.Name)
	}

	return result
}

// Get returns comment with given index
func (c Comments) Get(index int) *Comment {
	if len(c) == 0 || index >= len(c) {
		return nil
	}

	return c[index]
}

// Markdown converts comment HTML content to Markdown
func (c *Comment) Markdown() string {
	if c == nil || c.Content == "" {
		return ""
	}

	return htmlToMarkdown(c.Content)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// initEngine inits request engine
func initEngine() {
	if engine != nil {
		return
	}

	engine = &req.Engine{}
	engine.SetUserAgent(UA, "1")
}

// sendRequest sends request to API
func sendRequest(endpoint string, query req.Query, response any) error {
	resp, err := engine.Get(req.Request{
		URL:         apiURL + endpoint,
		Query:       query,
		Accept:      req.CONTENT_TYPE_JSON,
		AutoDiscard: true,
	})

	if err != nil {
		return fmt.Errorf("Can't send request to API: %w", err)
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("API returned non-ok status code %d", resp.StatusCode)
	}

	if response != nil {
		err = resp.JSON(response)

		if err != nil {
			return fmt.Errorf("Can't decode API response: %w", err)
		}
	}

	return nil
}

// convertIncidentsRequest converts incidents request to query
func convertIncidentsRequest(r IncidentsRequest) req.Query {
	q := req.Query{
		"lang":         strutil.Q(r.Lang, LANG_RU),
		"installation": strutil.Q(r.Region, "all"),
	}

	if !r.From.IsZero() {
		q["from"] = timeutil.Format(r.From, "%Y-%m-%d")
	}

	if !r.To.IsZero() {
		q["to"] = timeutil.Format(r.To, "%Y-%m-%d")
	}

	if r.Status != "" {
		q["status"] = r.Status
	}

	if len(r.Zones) > 0 {
		q["zones[]"] = r.Zones
	}

	return q
}

// ////////////////////////////////////////////////////////////////////////////////// //

// UnmarshalJSON parses JSON date
func (d *Date) UnmarshalJSON(b []byte) error {
	data := string(b)

	if data == "null" || data == `""` {
		d.Time = time.Time{}
		return nil
	}

	date, err := time.Parse(`"2006-01-02T15:04:05Z"`, data)

	if err != nil {
		return err
	}

	d.Time = date

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// htmlToMarkdown is simple html to markdown converter
func htmlToMarkdown(text string) string {
	if strings.Contains(text, "<img") {
		text, _ = reutil.Replace(htmlImgTagRegex, text, func(found string, submatch []string) string {
			return fmt.Sprintf("![IMG](%s)", submatch[0])
		})
	}

	if strings.Contains(text, "<a ") {
		text, _ = reutil.Replace(htmlLinkTagRegex, text, func(found string, submatch []string) string {
			return fmt.Sprintf("[%s](%s)", submatch[1], submatch[0])
		})
	}

	index := 0

	text, _ = reutil.Replace(htmlTagStartRegex, text, func(found string, submatch []string) string {
		switch submatch[0] {
		case "i":
			return "*"
		case "b", "strong":
			return "**"
		case "pre":
			return "`"
		case "code":
			return "```"
		case "ol":
			index = 1
			return ""
		case "ul":
			index = 0
			return ""
		case "br":
			return "\n"
		case "li":
			if index == 0 {
				return "• "
			} else {
				index += 1
				return fmt.Sprintf("%d. ", index-1)
			}
		}

		return ""
	})

	text, _ = reutil.Replace(htmlTagEndRegex, text, func(found string, submatch []string) string {
		switch submatch[0] {
		case "i":
			return "*"
		case "b", "strong":
			return "**"
		case "pre":
			return "`"
		case "code":
			return "```"
		}

		return ""
	})

	return html.UnescapeString(text)
}
