package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

const DbName = "screenshot.db"

type UserConfig struct {
	host string
	port int
	user string
	pass string
	path string
	link string
}

var db *sql.DB

func init() {
	conn, err := sql.Open("sqlite3", DbName)

	if err != nil {
		panic(err)
	}

	db = conn
}

func makeDbIfNotExist() {
	file, err := os.OpenFile(DbName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = file.Close()

	if err != nil {
		panic(err)
	}

	userConfigSql := `
create table if not exists userConfig
(
    host      varchar(255) not null,
    port      int          not null,
    user      varchar(255) not null,
    pass      varchar(255) not null,
    path      varchar(255) not null,
    link      varchar(255) not null,
    createdAt datetime     not null
);
create unique index if not exists userConfigUniq on userConfig (host, port, user, pass, path, link);
`

	statement, err := db.Prepare(userConfigSql)

	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = statement.Exec()

	if err != nil {
		panic(err)
	}
}

func getUserConfig() *UserConfig {
	stmt, err := db.Prepare(`
select
	uc.host,
	uc.port,
	uc.user,
	uc.pass,
	uc.path,
	uc.link
from userConfig uc
order by uc.createdAt desc
limit 1
`)

	if err != nil {
		panic(err)
	}

	rows, err := stmt.Query()

	if err != nil {
		panic(err)
	}

	defer func() {
		err = rows.Close()

		if err != nil {
			panic(err)
		}
	}()

	for rows.Next() {
		var host string
		var port int
		var user string
		var pass string
		var path string
		var link string

		err = rows.Scan(&host, &port, &user, &pass, &path, &link)

		return &UserConfig{host, port, user, pass, path, link}
	}

	return nil
}

func setUserConfig(host string, port int, user string, pass string, path string, link string) {
	_, err := db.Exec(
		"insert or ignore into userConfig (host, port, user, pass, path, link, createdAt) values ($1, $2, $3, $4, $5, $6, $7)",
		host,
		port,
		user,
		pass,
		path,
		link,
		time.Now().Format(time.RFC3339),
	)

	if err != nil {
		panic(err)
	}
}
