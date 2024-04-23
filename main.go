package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	_ "github.com/lib/pq"

	"github.com/BurntSushi/toml"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

const (
	// max and min count of status per day
	max = 20
	min = 1
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBname   string
}

type Count struct {
	new, done int
}

func main() {
	var dbConf DBConfig
	_, err := toml.DecodeFile("config.toml", &dbConf)

	host, port, user, password, dbname := dbConf.Host, dbConf.Port, dbConf.User, dbConf.Password, dbConf.DBname
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)
	fmt.Println("Established a successful connection!")

	// clear the table before every using
	clearString := `truncate table tickets`
	_, err = db.Exec(clearString)
	CheckError(err)

	year, month, day := time.Now().Date()

	// fill the table with random data from the last 7 days
	for i := day - 6; i <= day; i++ {
		randomStatusCount := rand.Intn(max-min) + min
		for j := 0; j < randomStatusCount; j++ {
			theme := GenerateTheme(10)
			timeCreation := GenerateDate(i, month, year)
			randomStatus := GenerateRandomStatus()

			insertDynString := `insert into "tickets" ("theme", "time_creation", "current_status") values ($1, $2, $3)`
			_, err = db.Exec(insertDynString, theme, timeCreation, randomStatus)
			CheckError(err)
		}
	}

	tickets := make(map[string]Count)

	rows, err := db.Query(`SELECT * FROM "tickets"`)
	CheckError(err)

	for rows.Next() {
		var theme string
		var dateCreation time.Time
		var status string

		err = rows.Scan(&theme, &dateCreation, &status)
		CheckError(err)

		dateCreationFormat := dateCreation.Format("02-January-2006")
		entry, ok := tickets[dateCreationFormat]
		if ok {
			if status == "new" {
				entry.new += 1
				tickets[dateCreationFormat] = entry
			} else if status == "done" {
				entry.done += 1
				tickets[dateCreationFormat] = entry
			}
		} else {
			if status == "new" {
				entry.new = 1
				tickets[dateCreationFormat] = entry
			} else if status == "done" {
				entry.done = 1
				tickets[dateCreationFormat] = entry
			}
		}
	}

	keys := SortedDates(tickets)
	CreateLineChart(tickets, keys)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func GenerateTheme(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	themeRunes := make([]rune, length)
	for index := range themeRunes {
		themeRunes[index] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(themeRunes)
}

func GenerateDate(day int, month time.Month, year int) string {
	var timeCreation string
	// negative day can be obtained if the current day in the month is less than 7 and there is getting to the previous month
	if day <= 0 {
		timeCreation = fmt.Sprint(day+CheckMonth(time.Month(int(month)-1))) + time.Month(int(month)-1).String() + fmt.Sprint(year)
	} else {
		timeCreation = fmt.Sprint(day) + month.String() + fmt.Sprint(year)
	}
	return timeCreation
}

func GenerateRandomStatus() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Int()%2 == 0 {
		return "done"
	} else {
		return "new"
	}
}

func CheckMonth(month time.Month) int {
	var result int
	switch int(month) {
	case 1, 3, 5, 7, 8, 10, 12:
		result = 31
	case 4, 6, 9, 11:
		result = 30
	case 2:
		result = 28
	}
	return result
}

func GenerateLineItems(tickets map[string]Count, keys []string, ticketStatus string) []opts.LineData {
	items := make([]opts.LineData, 0)

	for _, key := range keys {
		if ticketStatus == "new" {
			items = append(items, opts.LineData{Value: tickets[key].new})
		} else if ticketStatus == "done" {
			items = append(items, opts.LineData{Value: tickets[key].done})
		}
	}
	return items
}

func CreateLineChart(tickets map[string]Count, keys []string) {
	// create a new line instance
	line := charts.NewLine()

	// set some global options
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeInfographic,
			Width: "1200px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Продуктивность отдела"}),
	)

	// Put data into instance
	line.SetXAxis(keys).
		AddSeries("Выполненные статусы", GenerateLineItems(tickets, keys, "done"), charts.WithAreaStyleOpts(opts.AreaStyle{Color: "#c1232b"})).
		AddSeries("Новые статусы", GenerateLineItems(tickets, keys, "new"), charts.WithAreaStyleOpts(opts.AreaStyle{Color: "#27727b"})).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true, ShowSymbol: true, Symbol: "circle", SymbolSize: 10}))

	file, err := os.Create("the-efficiency-list.html")
	CheckError(err)

	err = line.Render(file)
	CheckError(err)
}

// sort dates in the map because the keys in the map aren't in order
func SortedDates(tickets map[string]Count) []string {
	keys := make([]string, len(tickets))
	i := 0
	for k := range tickets {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool {
		var result bool
		date1, _ := time.Parse("02-January-2006", keys[i])
		date2, _ := time.Parse("02-January-2006", keys[j])

		if int(date1.Month()) < int(date2.Month()) {
			result = true
		} else if int(date1.Month()) > int(date2.Month()) {
			result = false
		} else if int(date1.Day()) < int(date2.Day()) {
			result = true
		} else {
			result = false
		}
		return result
	})
	return keys
}
