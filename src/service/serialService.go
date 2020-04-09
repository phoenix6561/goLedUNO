package service

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tarm/serial"
)

type flusher interface {
	Flush() error
}

type commandError struct {
	Command string
	Arg     string
	Message string
}

func (e *commandError) Error() string {
	return fmt.Sprintf("command=%s;arg=%s;message=%s", e.Command, e.Arg, e.Message)
}

type thing struct {
	r *bufio.Reader
	w io.Writer
}

func newThing(rw io.ReadWriter) *thing {
	return &thing{w: rw, r: bufio.NewReader(rw)}
}

func (t *thing) runCommand(command, arg string) (string, error) {
	data := command
	if arg != "" {
		data += "=" + arg
	}
	var err error
	// windows ?
	//time.Sleep(2 * time.Second)
	log.Println("write")
	_, err = fmt.Fprint(t.w, data+"\n")
	if err != nil {
		return "", err
	}

	if f, ok := t.w.(flusher); ok {
		log.Println("flush")
		err = f.Flush()
		if err != nil {
			return "", err
		}
	}
	/*
		var response []string
		for {
			log.Println("read")
			line, err := t.r.ReadString('\n')
			if err != nil {
				return "", err
			}
			line = strings.TrimSpace(line)
			log.Println("line", line)

			if strings.HasPrefix(line, "ERR:") {
				return "", &commandError{
					Command: command,
					Arg:     arg,
					Message: strings.TrimSpace(strings.TrimPrefix(line, "ERR:")),
				}
			}

			if line == "OK" {
				break
			}

			response = append(response, line)
		}
	*/
	var response []string
	response = append(response, "OK")
	return strings.Join(response, "\n"), nil
}

//SendOverPort sends http request over serial port
func SendOverPort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081/")

	pathParams := mux.Vars(r)
	var portName string = "/dev/"
	if isWindows() {
		log.Println("is windows setting port to not use dev")
		portName = ""
	}

	var command string
	var args string
	var err error
	if val, ok := pathParams["port"]; ok {
		portName = portName + val
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a port"}`))
			return
		}
	}

	log.Println("port:" + portName)

	if val, ok := pathParams["command"]; ok {
		command = val
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a command"}`))
			return
		}
	}

	if val, ok := pathParams["args"]; ok {
		args = val
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a arg"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	port, err := serial.OpenPort(&serial.Config{Name: portName, Baud: 9600})

	if err != nil {
		log.Println("was unable to open port")
		return
	}

	defer port.Close()

	t := newThing(port)

	_, err = t.runCommand(command, args)
	if err != nil {
		log.Println(err)
	}

	//json.NewEncoder(w).Encode(books)
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
