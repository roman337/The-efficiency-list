package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/BurntSushi/toml"

	_ "github.com/lib/pq"
)

const (
	host = "localhost"
	port = 5432
	user = "postgres"
	// max and min count of status per day
	max = 20
	min = 1
)

type DBConfig struct {
	DBname   string
	Password string
}

type Count struct {
	new, done int
}

func main() {
	var dbConf DBConfig
	_, err := toml.DecodeFile("config.toml", &dbConf)

	password, dbname := dbConf.Password, dbConf.DBname
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", postgresqlDbInfo)
	CheckError(err)

	defer db.Close()
	err = db.Ping()
	CheckError(err)
	fmt.Println("Established a successful connection!")

	//rows, err := db.Query("select * from tickets")
	//CheckError(err)
	//fmt.Println(rows)
	// selectRow := `select * from tickets`

	deleteStr := `truncate table tickets`
	_, err = db.Exec(deleteStr)
	CheckError(err)

	year, month, day := time.Now().Date()
	for i := day - 6; i <= day; i++ {
		generateDoneCount := rand.Intn(max-min) + min
		//generateNewCount := rand.Intn(max-min) + min

		for j := 0; j < generateDoneCount; j++ {
			letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			themeRunes := make([]rune, 10)
			for index := range themeRunes {
				themeRunes[index] = letterRunes[rand.Intn(len(letterRunes))]
			}
			theme := string(themeRunes)
			timeCreation := fmt.Sprint(i) + month.String() + fmt.Sprint(year)
			randomStatus := RandomStatus()
			insertDynStr := `insert into "tickets" ("theme", "time_creation", "current_status") values ($1, $2, $3)`
			_, err = db.Exec(insertDynStr, theme, timeCreation, randomStatus)
			CheckError(err)
		}
	}

	rows, err := db.Query(`SELECT * FROM "tickets"`)
	fmt.Println(rows)

	var tickets = map[string]Count{}

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

	fmt.Print(tickets)
	CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func RandomStatus() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Int()%2 == 0 {
		return "done"
	} else {
		return "new"
	}
}
