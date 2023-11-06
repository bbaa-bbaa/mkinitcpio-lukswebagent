package main

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
)

type unlockRequest struct {
	Password string `json:"password"`
}

//go:embed static
var static embed.FS

var sysSignals = make(chan os.Signal, 8)
var requestPasswordSignals = make(chan int, 2)
var submitPassword = make(chan string)

func unlockLuksDevice() {
	systemdPasswordAgent := exec.Command("systemd-tty-ask-password-agent", "--query", "--watch")
	agent, _ := pty.Start(systemdPasswordAgent)
	defer agent.Close()
	p := make([]byte, 128)
	buf := make([]byte, 256)
	for {
		n, err := agent.Read(p)
		if err != nil {
			break
		}
		copy(buf, buf[n:])
		copy(buf[len(buf)-n:], p[:n])
		agentOutput := string(buf)
		if strings.Contains(agentOutput, "press TAB for no echo") {
			log.Println("luksAgent: Request for a password")
			requestPasswordSignals <- 1
			log.Println("luksAgent: Waiting for a password")
			password, ok := <-submitPassword
			if !ok {
				agent.Close()
				break
			}
			log.Printf("luksAgent: Get a password 「%s」", password)
			buf = make([]byte, 256)
			agent.WriteString(password)
			agent.Write([]byte{'\n'})
			log.Printf("luksAgent: Try password 「%s」", password)
		}
	}
}

func loadCerts(cfg *tls.Config) {
	certs, err := os.ReadDir("/etc/lukswebagent/certificates")
	if err != nil {
		return
	}
	for _, files := range certs {
		name := files.Name()
		if filepath.Ext(name) == ".crt" {
			cert, err := tls.LoadX509KeyPair("/etc/lukswebagent/certificates/"+name, "/etc/lukswebagent/certificates/"+name[:len(name)-len(filepath.Ext(name))]+".key")
			if err == nil {
				cfg.Certificates = append(cfg.Certificates, cert)
			}
		}
	}
}

func FileServerWith404(root http.FileSystem) http.Handler {
	// https://gist.github.com/lummie/91cd1c18b2e32fa9f316862221a6fd5c
	fs := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		upath = path.Clean(upath)

		f, err := root.Open(upath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		if err == nil {
			f.Close()
		}

		fs.ServeHTTP(w, r)
	})
}

func main() {
	log.SetPrefix("[LUKSWebAgent]")
	sub, _ := fs.Sub(static, "static")
	mux := &http.ServeMux{}
	mux.Handle("/", FileServerWith404(http.FS(sub)))
	mux.HandleFunc("/unlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, "{\"code\": 403, \"error\": \"method not allowed\"}")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{\"code\": 400, \"error\": \"Bad request\"}")
			return
		}
		var reqBody unlockRequest
		err = json.Unmarshal(body, &reqBody)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{\"code\": 400, \"error\": \"Bad request\"}")
			return
		}
		log.Printf("http: Wait for Request Password\r\n")
		<-requestPasswordSignals
		log.Printf("http: Get password 「%s」and send to agent to unlock luks device\r\n", reqBody.Password)
		submitPassword <- reqBody.Password
		log.Printf("http: Wait for unlock result\r\n")
		x, ok := <-requestPasswordSignals
		if x == 0 && ok {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{\"code\": 200}")
			log.Printf("http: Success unlock luks device\r\n")
			requestPasswordSignals <- 0
		} else {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{\"code\": 401, \"error\": \"Authenticate failed\"}")
			log.Printf("http: Failed to unlock luks device\r\n")
			requestPasswordSignals <- 2
		}
	})
	signal.Notify(sysSignals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		unlockLuksDevice()
	}()
	cfg := &tls.Config{}
	loadCerts(cfg)
	s := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: cfg,
	}
	go func() {
		s.ListenAndServeTLS("", "")
	}()

	<-sysSignals
	if len(requestPasswordSignals) == 0 {
		requestPasswordSignals <- 0
		<-requestPasswordSignals
	}
	close(requestPasswordSignals)
	close(submitPassword)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	s.Shutdown(ctx)
}
