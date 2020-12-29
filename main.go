package main

import (
	"bytes"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

func getVMs() (vncZones []string) {
	var (
		buf       *bytes.Buffer
		cmd       *exec.Cmd
		cmdOutput []byte
		allZones  []string
		line      string
		e         error
	)

	cmd = exec.Command("zoneadm", "list", "-np")

	if cmdOutput, e = cmd.Output(); e != nil {
		log.Println(e)
		return []string{}
	}

	buf = bytes.NewBuffer(cmdOutput)
	for line, e = buf.ReadString('\n'); e == nil; line, e = buf.ReadString('\n') {
		allZones = append(allZones, strings.TrimSuffix(line, "\n"))
	}

	for _, z := range allZones {
		var fields []string

		fields = strings.Split(z, ":")

		if fields[5] == "kvm" || fields[5] == "bhyve" {
			vncZones = append(vncZones, fields[1])
		}
	}

	return
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

	remote, e = net.Dial("unix", filepath.Join("/zones", vm, "root/tmp/vm.vnc"))
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(e)
		return
	}
	defer remote.Close()

	conn, e = upg.Upgrade(w, r, nil)
	if e != nil {
		log.Println(e)
		return
	}
	defer conn.Close()

	go toClient(conn, remote)
	fromClient(conn, remote)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/vnc/", http.FileServer(assetFS()))

	log.Fatal(http.ListenAndServe("127.0.0.1:8200", nil))
}
