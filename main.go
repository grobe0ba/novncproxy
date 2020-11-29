package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/grobe0ba/novncproxy/vm"
)

func getVMs() []vm.VM {
	var (
		buf        *bytes.Buffer
		cmd        *exec.Cmd
		cmdOutput  []byte
		vmUUIDList []string
		vmList     []vm.VM
		line       string
		e          error
	)

	cmd = exec.Command("vmadm", "lookup", "state=running", "brand=kvm")

	cmdOutput, e = cmd.Output()
	if e != nil {
		log.Fatal(e)
	}

	buf = bytes.NewBuffer(cmdOutput)
	for line, e = buf.ReadString('\n'); e == nil; line, e = buf.ReadString('\n') {
		vmUUIDList = append(vmUUIDList, strings.TrimSuffix(line, "\n"))
	}

	cmd = exec.Command("vmadm", "lookup", "state=running", "brand=bhyve")

	cmdOutput, e = cmd.Output()
	if e != nil {
		log.Fatal(e)
	}

	buf = bytes.NewBuffer(cmdOutput)
	for line, e = buf.ReadString('\n'); e == nil; line, e = buf.ReadString('\n') {
		vmUUIDList = append(vmUUIDList, strings.TrimSuffix(line, "\n"))
	}

	for _, VM := range vmUUIDList {
		var nVM [2]vm.VM

		cmd = exec.Command("vmadm", "info", VM)
		cmdOutput, e = cmd.Output()
		if e != nil {
			log.Fatal(e)
		}

		nVM[0] = vm.FromJSON(cmdOutput)

		cmd = exec.Command("vmadm", "get", VM)
		cmdOutput, e = cmd.Output()
		if e != nil {
			log.Fatal(e)
		}

		nVM[1] = vm.FromJSON(cmdOutput)
		nVM[0].UUID = nVM[1].UUID
		nVM[0].Alias = nVM[1].Alias

		vmList = append(vmList, nVM[0])
	}

	return vmList
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var (
		buf           *bytes.Buffer
		templateBytes []byte
		t             *template.Template
		vmList        []vm.VM
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
		host   string
		port   string
		rdr    io.Reader
		wrt    io.WriteCloser
		e      error
	)

	upg.ReadBufferSize = 1
	upg.WriteBufferSize = 1

	if port = r.FormValue("port"); port == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Connection port not specified by client " + r.URL.String())
		return
	}

	if host = r.FormValue("host"); host == "" {
		host = r.Host
	}

	conn, e = upg.Upgrade(w, r, nil)
	if e != nil {
		log.Println(e)
		return
	}
	defer conn.Close()

	remote, e = net.Dial("tcp", net.JoinHostPort(host, port))
	if e != nil {
		log.Println(e)
		return
	}
	defer remote.Close()

	_, rdr, e = conn.NextReader()
	if e != nil {
		log.Println(e)
		return
	}

	wrt, e = conn.NextWriter(websocket.BinaryMessage)
	if e != nil {
		log.Println(e)
		return
	}
	defer wrt.Close()

	go cat(rdr, remote)
	cat(remote, wrt)
}

func main() {
	var (
		certFile *string
		keyFile  *string
	)

	certFile = flag.String("cert", "novncproxy.pem", "TLS certificate PEM file")
	keyFile = flag.String("key", "novncproxy-key.pem", "TLS key PEM file")
	flag.Parse()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/vnc/", http.FileServer(assetFS()))

	log.Fatal(http.ListenAndServeTLS(":443", *certFile, *keyFile, nil))
}
