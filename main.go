package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	_ "github.com/lib/pq"
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
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", postgresqlDbInfo)
	CheckError(err)

	defer db.Close()
	err = db.Ping()
	CheckError(err)
	fmt.Println("Established a successful connection!")

	deleteStr := `truncate table tickets` // remove all datas before every using
	_, err = db.Exec(deleteStr)
	CheckError(err)

	year, month, day := time.Now().Date() // today
	for i := day - 6; i <= day; i++ {     //fill the table with random data from the last 7 days
		randomStatusCount := rand.Intn(max-min) + min
		for j := 0; j < randomStatusCount; j++ {
			letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			themeRunes := make([]rune, 10)
			for index := range themeRunes {
				themeRunes[index] = letterRunes[rand.Intn(len(letterRunes))]
			}
			theme := string(themeRunes)
			var timeCreation string
			if i <= 0 {
				timeCreation = fmt.Sprint(i+CheckMonth(time.Month(int(month)-1))) + time.Month(int(month)-1).String() + fmt.Sprint(year)
			} else {
				timeCreation = fmt.Sprint(i) + month.String() + fmt.Sprint(year)
			}
			randomStatus := RandomStatus()
			insertDynStr := `insert into "tickets" ("theme", "time_creation", "current_status") values ($1, $2, $3)`
			_, err = db.Exec(insertDynStr, theme, timeCreation, randomStatus)
			CheckError(err)
		}
	}

	rows, err := db.Query(`SELECT * FROM "tickets"`)

	tickets := make(map[string]Count)

	for rows.Next() {
		var theme string
		var dateCreation time.Time
		var status string

		//var doneCount, newCount = 0, 0

		err = rows.Scan(&theme, &dateCreation, &status)
		CheckError(err)
		dateCreationFormat := dateCreation.Format("02-January-2006")
		//fmt.Print(dateCreationFormat)
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

		//tickets[dateCreation.Format("02-January-2006")] = Count{newCount, doneCount}
		fmt.Println(theme, dateCreationFormat, status)
	}

	// keys := make([]string, 0, len(tickets))
	// for key, _ := range tickets {
	// 	keys = append(keys, key)
	// }
	// sort.Strings(keys)
	keys := sortedDates(tickets)
	fmt.Print(tickets)
	CheckError(err)
	createLineChart(tickets, keys)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
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

func RandomStatus() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Int()%2 == 0 {
		return "done"
	} else {
		return "new"
	}
}

func generateLineItems(tickets map[string]Count, keys []string, ticketStatus string) []opts.LineData {
	items := make([]opts.LineData, 0)
	if ticketStatus == "new" {
		for _, key := range keys {
			items = append(items, opts.LineData{Value: tickets[key].new})
			fmt.Println(key)
		}
	} else if ticketStatus == "done" {
		for _, key := range keys {
			items = append(items, opts.LineData{Value: tickets[key].done})
		}
	}
	return items
}

func createLineChart(tickets map[string]Count, keys []string) {
	// create a new line instance
	line := charts.NewLine()

	// set some global options like Title/Legend/ToolTip or anything else
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
		AddSeries("Новые статусы", generateLineItems(tickets, keys, "new")).
		AddSeries("Выполненные статусы", generateLineItems(tickets, keys, "done")).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	f, _ := os.Create("line.html")
	_ = line.Render(f)
}

func sortedDates(tickets map[string]Count) []string {
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
