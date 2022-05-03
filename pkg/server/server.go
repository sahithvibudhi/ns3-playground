package server

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sahithvibudhi/ns3-playground/config"
)

type WafRequestBody struct {
	Code string `json:"code"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func execCommand(l string) string {
	cmd := exec.Command("/bin/sh", "-c", "sudo "+l)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return stderr.String()
	}

	return out.String()
}

func saveCode(token, code string) {
	os.MkdirAll("uploads/"+token, 0755)
	f, err := os.Create("uploads/" + token + "/code")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err2 := f.WriteString(code)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func Start() {
	router := mux.NewRouter()

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.HandleFunc("/waf", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		t := WafRequestBody{}
		err := decoder.Decode(&t)
		if err != nil {
			panic(err)
		}

		token := RandStringRunes(8)

		var output string

		// save code to a file in uploads
		saveCode(token, t.Code)

		// run the code in a docker container
		output += execCommand(fmt.Sprintf("docker run -t -d --name ns3-%s ns3", token))

		// copy the code to docker container
		output += execCommand(fmt.Sprintf("docker cp ${PWD}/uploads/%s/code ns3-%s:/usr/ns-allinone-3.30.1/ns-3.30.1/scratch/file.cc", token, token))

		// compile the cpp using ./waf in ns3
		output += execCommand(fmt.Sprintf("docker exec ns3-%s sh -c \"cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf\"", token))

		// run the compiled file
		output += execCommand(fmt.Sprintf("docker exec ns3-%s sh -c \"cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf --run file\"", token))

		// copy pcap to output dir
		output += execCommand(fmt.Sprintf("docker exec ns3-%s bash -c \"mkdir -p /output; cp -f /usr/ns-allinone-3.30.1/ns-3.30.1/*.pcap /output\"", token))

		// copy the pcap files back to the host
		output += execCommand(fmt.Sprintf("docker cp ns3-%s:/output/. ${PWD}/uploads/%s/.", token, token))

		execCommand(fmt.Sprintf("docker stop ns3-%s", token))

		execCommand(fmt.Sprintf("dcoker rm ns3-%s", token))

		// count pcap files
		files, _ := ioutil.ReadDir(fmt.Sprintf("uploads/%s/", token))
		count := 0
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".pcap") {
				count++
			}
		}

		json.NewEncoder(w).Encode(map[string]string{"token": token, "pcapCount": fmt.Sprint(count), "output": output})
	})

	router.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		buf := new(bytes.Buffer)
		writer := zip.NewWriter(buf)

		files, _ := ioutil.ReadDir(fmt.Sprintf("uploads/%s/", token))
		for _, file := range files {
			filename := file.Name()
			data, err := ioutil.ReadFile(fmt.Sprintf("uploads/%s/%s", token, filename))
			if err != nil {
				log.Fatal(err)
			}
			f, err := writer.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			_, err = f.Write([]byte(data))
			if err != nil {
				log.Fatal(err)
			}
		}
		err := writer.Close()
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "ns3-playground"))
		//io.Copy(w, buf)
		w.Write(buf.Bytes())
	})

	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	root := filepath.Join(filepath.Dir(b), "../..", "./static/")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(root)))

	srv := &http.Server{
		Handler:      router,
		Addr:         config.Config.Server.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
