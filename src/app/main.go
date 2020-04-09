package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/jfox/restapi/src/dao"
	"github.com/jfox/restapi/src/service"
	"gopkg.in/yaml.v2"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
)

var db *sql.DB

type config struct {
	Database struct {
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
		Address  string `yaml:"address"`
		Scema    string `yaml:"scema"`
		Db       string `yaml:"db"`
	} `yaml:"database"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Mode struct {
		UseDb  string `yaml:"usedb"`
		HostFe string `yaml:"hostfe"`
	} `yaml:"mode"`
}

func getAllCommands(w http.ResponseWriter, r *http.Request) {

	var commands = dao.FindAll(db)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commands)

}
func getAllCommandsNoDB(w http.ResponseWriter, r *http.Request) {

	var commands []dao.Command
	commands = append(commands, *dao.NewCommand(1, "fake", "fake", "fake", "fake"))
	commands = append(commands, *dao.NewCommand(2, "fake2", "fake2", "fake2", "fake2"))
	commands = append(commands, *dao.NewCommand(3, "fake3", "fake3", "fake3", "fake3"))
	commands = append(commands, *dao.NewCommand(4, "fake4", "fake4", "fake4", "fake4"))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commands)

}

func main() {
	config := getConfig()

	r := mux.NewRouter()

	r.HandleFunc("/api/serial/{port}/{command}/{args}", service.SendOverPort).Methods("POST")

	if config.Mode.UseDb == "true" {
		r.HandleFunc("/api/commands", getAllCommands).Methods("GET")

		db = dao.Connect(config.Database.Db, config.Database.Username, config.Database.Password, config.Database.Address, config.Database.Scema)
		dao.Ping(db)
		fmt.Println("datadase address: " + config.Database.Address)
		fmt.Println("datadase user: " + config.Database.Username)
		fmt.Println("datadase type: " + config.Database.Db)
		fmt.Println("datadase scema: " + config.Database.Scema)

	} else {
		log.Println("not useing database returning fake data")
		r.HandleFunc("/api/commands", getAllCommandsNoDB).Methods("GET")

	}
	if config.Mode.HostFe == "true" {

		log.Println("Hosting index.html")
		// set up a new box by giving it a (relative) path to a folder on disk:
		box := packr.NewBox("./templates")
		// Get the string representation of a file, or an error if it doesn't exist:
		_, err := box.FindString("index.html")
		if err != nil {
			log.Fatal(err)
		}

		r.Handle("/", http.FileServer(box))
	} else {

		log.Println("user interface webserver disabled!")

	}

	fmt.Println("application started on port " + config.Server.Port)

	log.Fatal(http.ListenAndServe(config.Server.Port, r))

}

func getConfig() *config {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("dir: " + dir)

	f, err := os.Open(dir + string(os.PathSeparator) + "config.yml")

	if err != nil {
		log.Fatalln("failed to process config")
	}
	defer f.Close()

	var cfg config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalln("failed to process config")
	}

	return &cfg
}
