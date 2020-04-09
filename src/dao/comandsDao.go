package dao

import (
	"database/sql"
	"log"

	/*
		aliasing its package qualifier to _ so none of its exported names are visible to our code.
	*/
	_ "github.com/go-sql-driver/mysql"
)

//Command struct
type Command struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Port    string `json:"port"`
	Command string `json:"command"`
	Parm    string `json:"parm"`
}

//NewCommand returns a new command struct
func NewCommand(id int, address string, port string, command string, parm string) *Command {
	return &Command{ID: id, Address: address, Port: port, Command: command, Parm: parm}
}

// Connect to database
func Connect(dbType string, userName string, password string, address string, scema string) *sql.DB {

	db, err := sql.Open(dbType, userName+":"+password+"@tcp("+address+")/"+scema)
	if err != nil {
		log.Fatal(err)
	}

	Ping(db)

	return db

}

// Ping connection
func Ping(db *sql.DB) {

	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

}

//FindAll list all commands in db
func FindAll(db *sql.DB) []Command {
	var commands []Command
	var (
		id      int
		address string
		port    string
		command string
		parm    string
	)
	rows, err := db.Query("select * from commands")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &address, &port, &command, &parm)
		if err != nil {
			log.Fatal(err)
		}

		commands = append(commands, *NewCommand(id, address, port, command, parm))
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return commands
}
