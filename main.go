package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gorilla/websocket"
)

const ENV = "NOVNCPROXY_DAEMON"
const UUID = "a7fad6e2-7484-415c-a254-b28f9218ae8e"

func getVMs() []string {
	var (
		zonesDir *os.File
		zones    []string
		e        error
	)

	if zonesDir, e = os.Open("/zones"); e != nil {
		log.Fatal(e)
	}

	if zones, e = zonesDir.Readdirnames(-1); e != nil {
		log.Fatal(e)
	}

	return zones
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var (
		buf           *bytes.Buffer
		templateBytes []byte
		t             *template.Template
		vmList        []string
		e             error
	)

	buf = bytes.NewBuffer([]byte{})

	templateBytes, e = Asset("assets/__index.html")
	if e != nil {
		log.Fatal(e)
	}

	t, e = template.New("index").Parse(string(templateBytes))
	if e != nil {
		log.Fatal(e)
	}

	vmList = getVMs()

	e = t.Execute(buf, vmList)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, e = w.Write(buf.Bytes())
	if e != nil {
		log.Fatal(e)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		upg    websocket.Upgrader
		conn   *websocket.Conn
		remote net.Conn
		vm     string
		e      error
	)

	upg.ReadBufferSize = 1024
	upg.WriteBufferSize = 1024

	if vm = r.FormValue("vm"); vm == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Connection port not specified by client " + r.URL.String())
		return
	}

	conn, e = upg.Upgrade(w, r, nil)
	if e != nil {
		log.Println(e)
		return
	}
	defer conn.Close()

	remote, e = net.Dial("unix", filepath.Join("/zones", vm, "root/tmp/vm.vnc"))
	if e != nil {
		log.Println(e)
		return
	}
	defer remote.Close()

	go toClient(conn, remote)
	fromClient(conn, remote)
}

func main() {
	var (
		logFileFlag *string
		logFile     *os.File
		certFile    *string
		keyFile     *string
		pa          *syscall.ProcAttr
		argv0       string
		e           error
	)

	if os.Getenv(ENV) != UUID {
		os.Stdin.Close()
		os.Stdout.Close()
		os.Stderr.Close()

		pa = &syscall.ProcAttr{Env: []string{fmt.Sprintf("%s=%s", ENV, UUID)}}

		if argv0, e = os.Executable(); e != nil {
			log.Fatal(e)
		}

		if _, e = syscall.ForkExec(argv0, os.Args, pa); e != nil {
			log.Fatal(e)
			os.Exit(0)
		}
	}

	certFile = flag.String("cert", "novncproxy.pem", "TLS certificate PEM file")
	keyFile = flag.String("key", "novncproxy-key.pem", "TLS key PEM file")
	logFileFlag = flag.String("log", "/tmp/novncproxy.log", "Log file")
	flag.Parse()

	if logFile, e = os.Open(*logFileFlag); e != nil {
		log.Fatal(e)
	}

	log.SetOutput(logFile)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/vnc/", http.FileServer(assetFS()))

	log.Fatal(http.ListenAndServeTLS(":443", *certFile, *keyFile, nil))
}
