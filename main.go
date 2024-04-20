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

	max = 20
	min = 1
)

type DBConfig struct {
	DBname   string
	Password string
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
	//insertStr := `insert into "tickets" ("theme", "time_creation", "current_status") values ('second scheme', '2024-03-13', 'new')`
	//_, err = db.Exec(insertStr)
	//CheckError(err)
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
			insertDynStr := `insert into "tickets" ("theme", "time_creation", "current_status") values ($1, $2, 'done')`
			_, err = db.Exec(insertDynStr, theme, timeCreation)
			CheckError(err)
		}
	}

	rows, err := db.Query(`SELECT * FROM "tickets"`)
	fmt.Println(rows)

	var doneCount, newCount int

	for rows.Next() {
		var theme string
		var dateCreation time.Time
		var status string

		err = rows.Scan(&theme, &dateCreation, &status)
		CheckError(err)
		if status == "done" {
			doneCount++
		} else if status == "new" {
			newCount++
		}
		fmt.Println(theme, dateCreation.Format("02-Jan-2006"), status)
	}
	CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
