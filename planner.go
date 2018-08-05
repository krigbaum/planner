package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

//Define structures to receive weather forecast from JSON
type current struct {
	Time                 uint    `json:"time"`                 //	1453402675,
	Summary              string  `json:"summary"`              //	"Rain",
	Icon                 string  `json:"icon"`                 //	"rain",
	NearestStormDistance uint    `json:"nearestStormDistance"` //	0,
	PrecipIntensity      float64 `json:"precipIntensity"`      //	0.1685,
	PrecipIntensityError float64 `json:"precipIntensityError"` //	0.0067,
	PrecipProbability    float64 `json:"precipProbability"`    //	1,
	PrecipType           string  `json:"precipType"`           //	"rain",
	Temperature          float64 `json:"temperature"`          //	48.71,
	ApparentTemperature  float64 `json:"apparentTemperature"`  //	46.93,
	Dewpoint             float64 `json:"dewPoint"`             //	47.7,
	Humidity             float64 `json:"humidity"`             //	0.96,
	WindSpeed            float64 `json:"windSpeed"`            //	4.64,
	WindBearing          int     `json:"windBearing"`          //	186,
	Visibility           float64 `json:"visibility"`           //	4.3,
	CloudCover           float64 `json:"cloudCover"`           //	0.73,
	Pressure             float64 `json:"pressure"`             //	1009.7,
	Ozone                float64 `json:"ozone"`                //	328.35
}

type dailyData struct {
	Time                          uint64  `json:"time"`        //	1453402675,
	Summary                       string  `json:"summary"`     //	"Rain",
	Icon                          string  `json:"icon"`        //	"rain",
	SunriseTime                   uint    `json:"sunriseTime"` //	1453391560,
	SunsetTime                    uint    `json:"sunsetTime"`  //	1453424361,
	MoonPhase                     float64 `json:"moonPhase"`   //	0.43
	PrecipIntensity               float64 `json:"precipIntensity"`
	PrecipitationIntensityMax     float64 `json:"precipIntensityMax"`
	PrecipitationIntensityMaxTime float64 `json:"precipIntensityMaxTime"`
	PrecipProbability             float64 `json:"precipProbability"`           //	1,
	PrecipType                    string  `json:"precipType"`                  //	"rain",
	TemperatureHigh               float64 `json:"temperatureHigh"`             //	41.42,
	TemperatureHighTime           uint    `json:"temperatureHighTime"`         //	1453417200
	TemperatureLow                float64 `json:"temperatureLow"`              //	41.42,
	TemperatureLowTime            uint    `json:"temperatureLowTime"`          //	1453417200
	ApparentTemperatureHigh       float64 `json:"apparentTemperatureHigh"`     //	46.93,
	ApparentTemperatureHighTime   float64 `json:"apparentTemperatureHighTime"` //	46.93,
	ApparentTemperatureLow        float64 `json:"apparentTemperatureLow"`      //	46.93,
	ApparentTemperatureLowTime    float64 `json:"apparentTemperatureLowTime"`  //	46.93,
	Dewpoint                      float64 `json:"dewPoint"`                    //	47.7,
	Humidity                      float64 `json:"humidity"`                    //	0.96,
	Pressure                      float64 `json:"pressure"`
	WindSpeed                     float64 `json:"windSpeed"` //	4.64,
	WindGust                      float64 `json:"windGust"`
	WindGustTime                  float64 `json:"windGustTime"`
	WindBearing                   int     `json:"windBearing"` //	186,
	CloudCover                    float64 `json:"cloudCover"`
	UVIndex                       float64 `json:"uvIndex"`
	UVIndexTime                   float64 `json:"uvIndexTime"`
	Visibility                    float64 `json:"visibility"`                 //	4.3,
	Ozone                         float64 `json:"ozone"`                      //	328.35
	TemperatureMin                float64 `json:"temperatureMin"`             //	41.42,
	TemperatureMinTime            uint    `json:"temperatureMinTime"`         //	1453417200
	TemperatureMax                float64 `json:"temperatureMax"`             //	41.42,
	TemperatureMaxTime            uint    `json:"temperatureMaxTime"`         //	1453417200
	ApparentTemperatureMin        float64 `json:"apparentTemperatureMin"`     //	46.93,
	ApparentTemperatureMinTime    float64 `json:"apparentTemperatureMinTime"` //	46.93,
	ApparentTemperatureMax        float64 `json:"apparentTemperatureMax"`     //	46.93,
	ApparentTemperatureMaxTime    float64 `json:"apparentTemperatureMaxTime"` //	46.93,
}

type daily struct {
	Summary string      `json:"summary"` //	"Rain for the hour.",
	Icon    string      `json:"icon"`    //	"rain",
	Data    []dailyData `json:"data"`
}

type alert struct {
	Title       string `json:"title"`       //	"Flood Watch for Mason, WA",
	Time        uint   `json:"time"`        //	1453375020,
	Expires     uint   `json:"expires"`     //	1453407300,
	Description string `json:"description"` //	"...FLOOD WATCH...\n",
	URL         string `json:"uri"`         //	"http:/..."
}

type darkskyForecast struct {
	Latitude  float64 `json:"latitude"`  //	40.47780682531368,
	Longitude float64 `json:"longitude"` //	-86.93875375799722,
	Timezone  string  `json:"timezone"`  //	"America/Indiana/Indianapolis",
	Current   current `json:"currently"`
	Daily     daily
	Alerts    []alert
	Offset    int `json:"offset"` //	-4
} // End of receiving structure for weather forecast

type wotdType struct {
	Word      string
	Pronounce string
	POS       string
	Defs      []string
}

type sound struct {
	wave string `xml: "wav"	json:	"wave"`
	wpr  string `xml:	"wpr"	json:	"wpr"`
}

type entry struct {
	ew        string `xml: "ew"	json: "word"`
	subj      string `xml: "subj"	json: "subject"`
	syllables string `xml: "hw"	json: "syllables"`
	sound     string `xml: "sound"	json:	"sound"`
	pronounce string `xml:	"pr"	json:	"pronounce"`
	pos       string `xml: "fl"	json:	"pos"`
}

// Define receiving structure for WOTD XML
type wotdFormat struct {
	entryList string   `xml: "entry_list"	json: "entryList"`
	Word      string   `xml: "ew" json: "word"`
	Pronounce string   `xml: "pr" json: "pronounce"`
	POS       string   `xml: "fl" json: "pos"`
	Defs      []string `xml: "dt" json: "def"`
}

//Define structure to receive WOTD from XML
type entryList struct {
	XMLName xml.Name `xml:"entry_list"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Entry   struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id,attr"`
		Ew   struct {
			Text string `xml:",chardata"`
		} `xml:"ew"`
		Subj struct {
			Text string `xml:",chardata"`
		} `xml:"subj"`
		Hw struct {
			Text string `xml:",chardata"`
		} `xml:"hw"`
		Sound struct {
			Text string `xml:",chardata"`
			Wav  struct {
				Text string `xml:",chardata"`
			} `xml:"wav"`
			Wpr struct {
				Text string `xml:",chardata"`
			} `xml:"wpr"`
		} `xml:"sound"`
		Pr struct {
			Text string `xml:",chardata"`
		} `xml:"pr"`
		Fl struct {
			Text string `xml:",chardata"`
		} `xml:"fl"`
		In []struct {
			Text string `xml:",chardata"`
			If   struct {
				Text string `xml:",chardata"`
			} `xml:"if"`
		} `xml:"in"`
		Et struct {
			Text string `xml:",chardata"`
			It   []struct {
				Text string `xml:",chardata"`
			} `xml:"it"`
		} `xml:"et"`
		Def struct {
			Text string `xml:",chardata"`
			Vt   struct {
				Text string `xml:",chardata"`
			} `xml:"vt"`
			Date struct {
				Text string `xml:",chardata"`
			} `xml:"date"`
			Sn []struct {
				Text string `xml:",chardata"`
			} `xml:"sn"`
			Dt []struct {
				Text string `xml:",chardata"`
				Sx   struct {
					Text string `xml:",chardata"`
					Sxn  struct {
						Text string `xml:",chardata"`
					} `xml:"sxn"`
				} `xml:"sx"`
				Vi struct {
					Text string `xml:",chardata"`
					It   struct {
						Text string `xml:",chardata"`
					} `xml:"it"`
				} `xml:"vi"`
			} `xml:"dt"`
		} `xml:"def"`
		Uro []struct {
			Text string `xml:",chardata"`
			Ure  struct {
				Text string `xml:",chardata"`
			} `xml:"ure"`
			Sound struct {
				Text string `xml:",chardata"`
				Wav  struct {
					Text string `xml:",chardata"`
				} `xml:"wav"`
				Wpr struct {
					Text string `xml:",chardata"`
				} `xml:"wpr"`
			} `xml:"sound"`
			Pr struct {
				Text string `xml:",chardata"`
			} `xml:"pr"`
			Fl struct {
				Text string `xml:",chardata"`
			} `xml:"fl"`
		} `xml:"uro"`
	} `xml:"entry"`
}

//Define structures to receive configuration from JSON
type configStruct struct {
	Debug                 bool
	DarkSkyKey            string
	Latitude              string
	Longitude             string
	Excludes              string
	WeatherURL            string
	WeatherReloadInterval int
	QotdURL               string
	QotdReloadInterval    int
	WotdURL               string
	WotdReloadInterval    int
	PhotosDir             string
	CSSDirectory          string
	PhotoReloadInterval   int
	TimeCheckInterval     int
	HTMLFile              string
	MWrss                 string
	MWurl                 string
	MWkey                 string
	MaxPlannerLog         int
	MaxWeatherLog         int
	MaxWOTDLog            int
	MaxPhotoLog           int
} // End of receiving structure for configuration

var forecast darkskyForecast

func main() {
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Starting Planner Application.\n")
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Loading Configuration from json/config.json.\n\n")

	config := getConfig()
	displayConfig(config)

	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Calling startWeather()\n")
	go startWeather(config)
	time.Sleep(10 * time.Second)

	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Calling startWOTD()\n")
	go startWOTD(config)
	time.Sleep(10 * time.Second)

	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Calling startPhotos()\n")
	go startPhotos(config)
	time.Sleep(10 * time.Second)

	logger("calendar", time.Now().Format(time.RFC850)+"  INFO: Calling startCalendar()\n")
	go startCalendar(config)
	select {}
}

func startWeather(config configStruct) {
	// Initial Weather load on startup
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Initial Weather() Load\n")
	getWeather(config)

	// Repeat Weather load every weatherdReloadInterval
	ticker := time.NewTicker(time.Hour * time.Duration(config.WeatherReloadInterval))
	for range ticker.C {
		logger("planner", time.Now().Format(time.RFC850)+"  INFO: Periodic Weather() Load\n")
		getWeather(config)
	}
	logger("planner", time.Now().Format(time.RFC850)+"\n  INFO: *** Error: Exit on range ticker in function startWeather(). ***\n\n")
}

func startWOTD(config configStruct) {
	// Initial WOTD load on startup
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Initial WOTD() Load\n")
	getWOTD(config)

	// Repeat WOTD load every wotdReloadInterval
	ticker := time.NewTicker(time.Hour * time.Duration(config.WotdReloadInterval))
	for range ticker.C {
		logger("planner", time.Now().Format(time.RFC850)+"  INFO: Periodic WOTD() Load\n")
		getWOTD(config)
	}
	logger("planner", time.Now().Format(time.RFC850)+"\n  INFO: *** Error: Exit on range ticker in function startWOTD(). ***\n\n")
}

func startPhotos(config configStruct) {
	// Initial Photos load on startup
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Initial Photos() Load\n")
	getPhotos(config)

	// Repeat WOTD load every wotdReloadInterval
	ticker := time.NewTicker(time.Minute * time.Duration(config.PhotoReloadInterval))
	for range ticker.C {
		logger("planner", time.Now().Format(time.RFC850)+"  INFO: Periodic Photos() Load\n")
		getPhotos(config)
	}
	logger("planner", time.Now().Format(time.RFC850)+"\n  INFO: *** Error: Exit on range ticker in function startPhotos(). ***\n\n")
}

func startCalendar(config configStruct) {
	// Initial Calendar load on startup
	logger("planner", time.Now().Format(time.RFC850)+"  INFO: Initial Calendar() Load\n")
	getCalendar(config)

	// Repeat Calendar load every 24 hours
	ticker := time.NewTicker(time.Hour * time.Duration(12))
	for range ticker.C {
		logger("planner", time.Now().Format(time.RFC850)+"  INFO: Periodic Calendar() Load\n")
		getCalendar(config)
	}
	logger("planner", time.Now().Format(time.RFC850)+"\n  INFO: *** Error: Exit on range ticker in function startCalendar(). ***\n\n")
}

func getPhotos(config configStruct) {
	cssBytes, err := ioutil.ReadFile(config.CSSDirectory)
	if err != nil {
		logger("photo", time.Now().Format(time.RFC850)+"  ERROR: ReadFile failed on "+config.CSSDirectory+"\n")
	}
	css := string(cssBytes)

	rand.Seed(time.Now().Unix())

	deck, err := ioutil.ReadDir(config.PhotosDir)
	if err != nil {
		logger("photo", time.Now().Format(time.RFC850)+"  INFO: ReadDir error on"+config.PhotosDir+"\n")
	}

	index := rand.Intn(len(deck))
	photo := deck[index].Name()

	startStr := "background: url(../photos/"
	stopStr := ") no-repeat center center fixed"
	start := strings.Index(css, startStr)
	stop := strings.Index(css, stopStr) + len(stopStr)
	oldStr := css[start:stop]
	newStr := startStr + photo + stopStr
	css = strings.Replace(css, oldStr, newStr, 1)

	cssFile := []byte(css)
	ioutil.WriteFile(config.CSSDirectory, cssFile, 0644)
}

func getWeather(config configStruct) {
	htmlBytes, err := ioutil.ReadFile(config.HTMLFile)
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"ReadFile failed w/ err on"+config.HTMLFile+"\n")
	}
	html := string(htmlBytes)

	darkskyURL := config.WeatherURL + config.DarkSkyKey + "/" + config.Latitude + "," + config.Longitude + "?" + config.Excludes
	forecast = getForecast(darkskyURL)
	forecast.Daily.Data = forecast.Daily.Data[:3]

	startStr := "<span id=\"currentTemp\">"
	stopStr := " &#8457"
	valueStr := string(truncate(forecast.Current.Temperature, 0))
	start := strings.Index(html, startStr)
	stop := strings.Index(html, stopStr) + len(stopStr)
	oldStr := html[start:stop]
	newStr := startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"currentHumidity\">"
	stopStr = " %</span>"
	valueStr = string(truncate(forecast.Current.Humidity*100, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"currentWindSpeed\">"
	stopStr = " mph</span>"
	valueStr = string(truncate(forecast.Current.WindSpeed, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"currentVisibility\">"
	stopStr = " mi.</span>"
	valueStr = string(truncate(forecast.Current.Visibility, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<h2><span id=\"day1\">"
	stopStr = "<!--d1--></span></h2>"
	valueStr = getWeekday(forecast.Daily.Data[0].Time)
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"lowTemp1\">"
	stopStr = " &#8457;<!--1--></span>"
	valueStr = string(truncate(forecast.Daily.Data[0].TemperatureLow, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"highTemp1\">"
	stopStr = " &#8457;<!--2--></span>"
	valueStr = string(truncate(forecast.Daily.Data[0].TemperatureHigh, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"humidity1\">"
	stopStr = " %<!--1--></span>"
	valueStr = string(truncate(forecast.Daily.Data[0].Humidity*100, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"windspeed1\">"
	stopStr = " mph<!--1--></span>"
	valueStr = string(truncate(forecast.Daily.Data[0].WindSpeed, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"visibility1\">"
	stopStr = " mi.<!--1--></span>"
	valueStr = string(truncate(forecast.Daily.Data[0].Visibility, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<h2><span id=\"day2\">"
	stopStr = "<!--d2--></span></h2>"
	valueStr = getWeekday(forecast.Daily.Data[1].Time)
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"lowTemp2\">"
	stopStr = " &#8457;<!--3--></span>"
	valueStr = string(truncate(forecast.Daily.Data[1].TemperatureLow, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"highTemp2\">"
	stopStr = " &#8457;<!--4--></span>"
	valueStr = string(truncate(forecast.Daily.Data[1].TemperatureHigh, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"humidity2\">"
	stopStr = " %<!--2--></span>"
	valueStr = string(truncate(forecast.Daily.Data[1].Humidity*100, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"windspeed2\">"
	stopStr = " mph<!--2--></span>"
	valueStr = string(truncate(forecast.Daily.Data[1].WindSpeed, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"visibility2\">"
	stopStr = " mi.<!--2--></span>"
	valueStr = string(truncate(forecast.Daily.Data[1].Visibility, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<h2><span id=\"day3\">"
	stopStr = "<!--d3--></span></h2>"
	valueStr = getWeekday(forecast.Daily.Data[2].Time)
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"lowTemp3\">"
	stopStr = " &#8457;<!--5--></span>"
	valueStr = string(truncate(forecast.Daily.Data[2].TemperatureLow, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"highTemp3\">"
	stopStr = " &#8457;<!--6--></span>"
	valueStr = string(truncate(forecast.Daily.Data[2].TemperatureHigh, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"humidity3\">"
	stopStr = " %<!--3--></span>"
	valueStr = string(truncate(forecast.Daily.Data[2].Humidity*100, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"windspeed3\">"
	stopStr = " mph<!--3--></span>"
	valueStr = string(truncate(forecast.Daily.Data[2].WindSpeed, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<br> <span id=\"visibility3\">"
	stopStr = " mi.<!--3--></span>"
	valueStr = string(truncate(forecast.Daily.Data[2].Visibility, 0))
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	htmlFile := []byte(html)
	ioutil.WriteFile(config.HTMLFile, htmlFile, 0644)

	logger("weather", time.Now().Format(time.RFC850)+"  INFO: Finished getWeather()\n")
}

func getCalendar(config configStruct) {
	var dateStr string
	var startStr string
	loop := 0

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	oauth2config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := calendar.New(getClient(oauth2config))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	htmlBytes, err := ioutil.ReadFile(config.HTMLFile)
	if err != nil {
		log.Fatalln("ReadFile failed w/ err", err)
	}
	html := string(htmlBytes)

	//layout := "2006-01-02T15:04:05Z"
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	logger("calendar", "Upcoming events:")
	if len(events.Items) == 0 {
		logger("calendar", "No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			loop++
			splitstr := strings.Split(date, "T")
			if len(splitstr) == 1 {
				datefmt := "2006-01-02"
				tm, err := time.Parse(datefmt, splitstr[0])
				if err != nil {
					errmsg := fmt.Sprintf("Error in time.Parse(): %s", err)
					logger("calendar", errmsg)
				}
				dateStr = tm.Format("Mon Jan 2")
			}

			if len(splitstr) == 2 {
				dtStr := splitstr[0] + " " + splitstr[1]
				dtStr = dtStr[:16]
				datefmt := "2006-01-02 15:04"
				tm, err := time.Parse(datefmt, dtStr)
				if err != nil {
					errmsg := fmt.Sprintf("Error in time.Parse(): %s", err)
					logger("calendar", errmsg)
				}
				dateStr = tm.Format("Monday Jan 2 at 3:04pm")
			}

			logger("calendar", item.Summary+" ("+dateStr+")")
			startStr = "<li id=\"item" + strconv.Itoa(loop) + "\">"
			stopStr := "<!-- e" + strconv.Itoa(loop) + " --></li>"
			valueStr := item.Summary + " (" + dateStr + ")"
			start := strings.Index(html, startStr)
			stop := strings.Index(html, stopStr) + len(stopStr)
			oldStr := html[start:stop]
			newStr := startStr + valueStr + stopStr
			html = strings.Replace(html, oldStr, newStr, 1)

		}
		htmlFile := []byte(html)
		ioutil.WriteFile(config.HTMLFile, htmlFile, 0644)
	}
}

func getConfig() configStruct {
	// Read config.json file and assign values to struct config ===================================
	var config configStruct
	configFile, err := ioutil.ReadFile("json/config.json")
	if err != nil {
		logger("planner", time.Now().Format(time.RFC850)+"  INFO: File error reading json/config.json\n")
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(configFile), &config)
	if err != nil {
		logger("planner", time.Now().Format(time.RFC850)+"  FATAL: Error unmarshaling json/config.json:")
	}

	return config
}

func displayConfig(config configStruct) {
	logger("planner", "                Debug: "+strconv.FormatBool(config.Debug)+"\n")
	logger("planner", "           darkSkyKey: "+config.DarkSkyKey+"\n")
	logger("planner", "             latitude: "+config.Latitude+"\n")
	logger("planner", "            longitude: "+config.Longitude+"\n")
	logger("planner", "             excludes: "+config.Excludes+"\n")

	logger("planner", "           weatherURL: "+config.WeatherURL+"\n")
	logger("planner", "weatherReloadInterval: "+strconv.Itoa(config.WeatherReloadInterval)+" Hr.\n")

	logger("planner", "              qotdURL: "+config.QotdURL+"\n")
	logger("planner", "   qotdReloadInterval: "+strconv.Itoa(config.QotdReloadInterval)+" Hr.\n")

	logger("planner", "              wotdURL: "+config.WotdURL+"\n")
	logger("planner", "   wotdReloadInterval: "+strconv.Itoa(config.WotdReloadInterval)+" Hr.\n")

	logger("planner", "            photosDir: "+config.PhotosDir+"\n")
	logger("planner", "         cssDirectory: "+config.CSSDirectory+"\n")
	logger("planner", "  photoReloadInterval: "+strconv.Itoa(config.PhotoReloadInterval)+" Min.\n")

	logger("planner", "    timeCheckInterval: "+strconv.Itoa(config.TimeCheckInterval)+" Sec.\n")

	logger("planner", "             HTMLFile: "+config.HTMLFile+"\n")

	logger("planner", "                mwRSS: "+config.MWrss+"\n")
	logger("planner", "                mwURL: "+config.MWurl+"\n")
	logger("planner", "                mwKEY: "+config.MWkey+"\n")

	logger("planner", "        maxPlannerLog: "+strconv.Itoa(config.MaxPlannerLog)+" M.\n")
	logger("planner", "        maxWeatherLog: "+strconv.Itoa(config.MaxWeatherLog)+" M.\n")
	logger("planner", "           maxWOTDLog: "+strconv.Itoa(config.MaxWOTDLog)+" M.\n")
	logger("planner", "          maxPhotoLog: "+strconv.Itoa(config.MaxPhotoLog)+" M.\n\n")
}

func getForecast(darkskyURL string) darkskyForecast {
	var forecast darkskyForecast

	_, err := os.Stat("json/darksky.json")
	if !os.IsNotExist(err) {
		err := os.Remove("json/darksky.json")
		if err != nil {
			logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Error removing json/darksky.json.\n")
			logger("weather", time.Now().Format(time.RFC850)+"   FATAL: Exiting program.\n")
			os.Exit(1)
		} else {
			logger("weather", time.Now().Format(time.RFC850)+"  INFO: json/darksky.json has been deleted.\n")
		}
	} else {
		logger("weather", time.Now().Format(time.RFC850)+"  INFO: json/darksky.json does not exist.\n")
	}
	logger("weather", time.Now().Format(time.RFC850)+"  INFO: json/darksky.json has been loaded.\n")

	data, err := http.Get(darkskyURL)
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"    FATAL: Error reading from darkSky\n")
		logger("weather", time.Now().Format(time.RFC850)+"    FATAL: Exiting program.\n")
		os.Exit(1)
	}

	// Convert raw data to []bytes.
	dataBYTES, err := ioutil.ReadAll(data.Body)
	data.Body.Close()
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Error reading body darksky\n")
		logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Exiting program.\n")
		os.Exit(1)
	}

	_, err = os.Stat("json/darksky.json")
	if !os.IsNotExist(err) {
		err := os.Remove("json/darksky.json")
		if err != nil {
			logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Error removing json/darksky.json.\n")
			logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Exiting program.\n")
			os.Exit(1)
		}
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, dataBYTES, "", "    ")
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"  INFO: Error pretty printing JSON\n")
	}

	darksky, err := os.OpenFile("json/darksky.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"  INFO: Error opening 'json/darksky.json\n")
	}
	defer darksky.Close()

	darksky.WriteString(prettyJSON.String())

	// Start of unmarshal
	weatherData, err := ioutil.ReadFile("json/darksky.json")
	err = json.Unmarshal(weatherData, &forecast)
	if err != nil {
		logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Error unmarshaling json/config.json\n")
		logger("weather", time.Now().Format(time.RFC850)+"  FATAL: Exiting program.\n")
		os.Exit(1)
	}

	logger("weather", time.Now().Format(time.RFC850)+"  INFO: Finished getForecastData()\n")
	return forecast
}

func getWOTD(config configStruct) {
	rssURL := config.MWrss
	data, err := http.Get(rssURL)
	if err != nil {
		logger("wotd", time.Now().Format(time.RFC850)+"  INFO: Error on http.Get(rssURL):\n")
	}
	dataBYTES, err := ioutil.ReadAll(data.Body)
	if err != nil {
		logger("wotd", time.Now().Format(time.RFC850)+"  INFO: Error on ioutil.ReadAll(data.Body):")
	}
	data.Body.Close()
	rss := string(dataBYTES)

	word := extract(rss, "<![CDATA[", "]]>")
	wotdURL := config.MWurl + word + "?key=" + config.MWkey
	data, err = http.Get(wotdURL)
	if err != nil {
		logger("wotd", time.Now().Format(time.RFC850)+"  INFO: Error on http.Get(wotdURL)")
	}
	dataBYTES, err = ioutil.ReadAll(data.Body)
	if err != nil {
		logger("wotd", time.Now().Format(time.RFC850)+"  INFO: Error on ioutil.ReadAll(data.Body):")
	}
	logger("wotd", time.Now().Format(time.RFC850)+"  dataBYTES ="+string(dataBYTES))
	data.Body.Close()
	logger("wotd", string(dataBYTES))

	//  For XML test only
	//d1 := dataBYTES
	//err = ioutil.WriteFile("dict.xml", d1, 0644)

	var def1 entryList
	err = xml.Unmarshal(dataBYTES, &def1)
	if err != nil {
		errmsg := fmt.Sprintf(time.Now().Format(time.RFC850)+"  INFO: Error on xml.Unmarshall(dataBytes): %s", err)
		logger("wotd", errmsg)
	}
	result := fmt.Sprintf("%+v\n", def1)
	logger("wotd", result)

	var wotdInfo wotdType
	logger("wotd", "XML = "+string(dataBYTES))
	wotdInfo.Word = def1.Entry.ID
	logger("wotd", time.Now().Format(time.RFC850)+"  Word: "+wotdInfo.Word+"\n")
	wotdInfo.Pronounce = def1.Entry.Pr.Text
	logger("wotd", time.Now().Format(time.RFC850)+"  Pronunciation: "+wotdInfo.Pronounce+"\n")
	wotdInfo.POS = def1.Entry.Fl.Text
	logger("wotd", time.Now().Format(time.RFC850)+"  Part of Speech: "+wotdInfo.POS+"\n")
	numdefs := len(def1.Entry.Def.Dt)
	logger("wotd", time.Now().Format(time.RFC850)+"  No. of Definitions: "+strconv.Itoa(numdefs)+"\n")
	x := 0

	for x < numdefs {
		if len(def1.Entry.Def.Dt[x].Text) > 0 {
			wotdInfo.Defs = append(wotdInfo.Defs, string(def1.Entry.Def.Dt[x].Text))
			logger("wotd", time.Now().Format(time.RFC850)+string(def1.Entry.Def.Dt[x].Text))
		}
		x++
	}

	htmlBytes, err := ioutil.ReadFile("planner.html")
	if err != nil {
		log.Fatalln("ReadFile failed w/ err", err)
	}
	html := string(htmlBytes)

	startStr := "<span id=\"word\">"
	stopStr := ":&nbsp;<!--w1--></span>"
	valueStr := string(wotdInfo.Word)
	start := strings.Index(html, startStr)
	stop := strings.Index(html, stopStr) + len(stopStr)
	oldStr := html[start:stop]
	newStr := startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"pronounce\">[&nbsp;"
	stopStr = "&nbsp;]<!--w2--></span>"
	valueStr = "&nbsp;" + string(wotdInfo.Pronounce)
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	startStr = "<span id=\"pos\">"
	stopStr = "<!--w3--></span>"
	valueStr = "&nbsp;" + string(wotdInfo.POS)
	start = strings.Index(html, startStr)
	stop = strings.Index(html, stopStr) + len(stopStr)
	oldStr = html[start:stop]
	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	valueStr = ""
	d := 0
	for d < len(wotdInfo.Defs) {
		startStr = "<span id=\"defs\">"
		stopStr = "<!--w4--></span>"
		cleanerdef := erase(string(wotdInfo.Defs[d]), ":")
		valueStr = valueStr + "&nbsp;&nbsp;&nbsp;Definition " + strconv.Itoa(d+1) + ") &nbsp;" + cleanerdef + "<br>"
		start = strings.Index(html, startStr)
		stop = strings.Index(html, stopStr) + len(stopStr)
		oldStr = html[start:stop]
		d++
	}

	newStr = startStr + valueStr + stopStr
	html = strings.Replace(html, oldStr, newStr, 1)

	htmlFile := []byte(html)
	ioutil.WriteFile("planner.html", htmlFile, 0644)

	logger("wotd", time.Now().Format(time.RFC850)+"  INFO: Finished getWOTD()\n")
}

func truncate(x interface{}, p int) string {
	fmtStr := "%." + strconv.Itoa(p) + "f"
	xstring := fmt.Sprintf(fmtStr, x)
	return xstring
}

func getWeekday(UnixTime uint64) string {
	timeStr := strconv.FormatUint(UnixTime, 10)
	reportedTime, _ := strconv.ParseInt(timeStr, 10, 64)
	tm := time.Unix(reportedTime, 0)
	day0 := fmt.Sprintf("%v", tm.Weekday())
	return day0
}

func getTime(reportedStr string) time.Time {
	reportedTime, _ := strconv.ParseInt(reportedStr, 10, 64)
	tm := time.Unix(reportedTime, 0)
	return tm
}

func extract(src string, startStr string, stopStr string) string {
	start := strings.Index(src, startStr) + len(startStr)
	if start == -1 {
		return "NotFound"
	}

	stop := strings.Index(src, stopStr)
	if stop == -1 {
		return "NotFound"
	}

	found := src[start:stop]
	return found
}

func erase(src string, ch string) string {
	if len(ch) > 1 || len(ch) == 0 {
		return "erase() failed on ch"
	}
	p := 0
	for p < len(src) {
		if string(src[p]) == ch {
			src = src[:p] + src[p+1:]
		}
		p++
	}
	return src
}

func logger(logname string, message string) {
	logName := "log/" + logname + ".log"
	bakName := "log/" + logname + ".bak"
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("File", logName, "failed with error:", err)
	}
	test, err := os.Stat(logName)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		fmt.Printf(errStr)
	}
	size := test.Size()
	if size > 2048 {
		fmt.Println(logName, "exceeds file size limit:", size)
		copyFile(logName, bakName)
		if err != nil {
			fmt.Print("CopyFile Error:", err)
		}
		err := os.Remove(logName)
		if err != nil {
			fmt.Println("Error removing", logName, ":", err)
		}
	}
	_, err = io.WriteString(f, message)
	if err != nil {
		fmt.Printf("Write failed with error:", err)
	}
}

func copyFile(src, dest string) (err error) {
	fmt.Println("From copyFile(): src =", src, "   dest=", dest)
	dat, err := ioutil.ReadFile(src)
	if err != nil {
		fmt.Println("err in ioutil.ReadFile():", err)
	}

	err = os.Remove(dest)
	if err != nil {
		fmt.Println("Unable to delete", dest)
	}

	f, err := os.Create(dest)
	if err != nil {
		fmt.Println("Unable to create", dest)
	}
	defer f.Close()

	f.WriteString(string(dat))
	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Fatalf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}
