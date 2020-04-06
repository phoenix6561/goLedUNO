package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

	return strings.Join(response, "\n"), nil
}

// Get all books
func sendOverPort(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	port, err := serial.OpenPort(&serial.Config{Name: "COM3", Baud: 9600})
	// options := serial.OpenOptions{
	// 	PortName: "COM3",
	// 	BaudRate: 115200,
	// 	DataBits: 8,
	// 	StopBits: 1,
	// }

	// Open the port.
	// port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	defer port.Close()

	t := newThing(port)

	_, err = t.runCommand("ATCOLOR", "B_ON")
	if err != nil {
		log.Fatalln(err)
	}

	state, err := t.runCommand("ATCOLOR?", "")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("state:", state)

	_, err = t.runCommand("INVALID", "")
	if err != nil {
		log.Println("expected error:", err)
	}

	//json.NewEncoder(w).Encode(books)
}

func main() {

	port, err := serial.OpenPort(&serial.Config{Name: "COM3", Baud: 9600})
	// options := serial.OpenOptions{
	// 	PortName: "COM3",
	// 	BaudRate: 115200,
	// 	DataBits: 8,
	// 	StopBits: 1,
	// }

	// Open the port.
	// port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()

	t := newThing(port)

	_, err = t.runCommand("ATBEEP", "")
	if err != nil {
		log.Fatalln(err)
	}

	state, err := t.runCommand("ATCOLOR?", "")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("state:", state)

	_, err = t.runCommand("INVALID", "")
	if err != nil {
		log.Println("expected error:", err)
	}

	// fmt.Println("application started on port 8000")
	// r := mux.NewRouter()

	// r.HandleFunc("/api/serial", sendOverPort).Methods("POST")

	// log.Fatal(http.ListenAndServe(":8000", r))

}
