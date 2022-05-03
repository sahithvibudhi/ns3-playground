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
	"time"

	"github.com/gorilla/mux"
	"github.com/sahithvibudhi/ns3-playground/config"
	"github.com/sahithvibudhi/ns3-playground/pkg/logger"
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
	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	root := filepath.Join(filepath.Dir(b), "../..", "uploads/"+token)
	os.MkdirAll(root, 0755)

	// Root folder of this project
	root = filepath.Join(filepath.Dir(b), "../..", "uploads/"+token+"/code")
	f, err := os.Create(root)
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
		logger.Logger.Println(fmt.Sprintf("Running ns3 for %s", token))

		// run the code in a docker container
		logger.Logger.Println("creating docker container")
		output += execCommand(fmt.Sprintf("docker run -t -d --name ns3-%s ns3", token))

		_, b, _, _ := runtime.Caller(0)

		// Root folder of this project
		code := filepath.Join(filepath.Dir(b), "../..", "uploads/"+token+"/code")
		// Root folder of this project
		requestRoot := filepath.Join(filepath.Dir(b), "../..", "uploads/"+token)

		// copy the code to docker container
		logger.Logger.Println("copying files to the docker container")
		output += execCommand(fmt.Sprintf("docker cp %s ns3-%s:/usr/ns-allinone-3.30.1/ns-3.30.1/scratch/file.cc", code, token))

		// compile the cpp using ./waf in ns3
		logger.Logger.Println("./waf configuring")
		output += execCommand(fmt.Sprintf("docker exec ns3-%s sh -c \"cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf configure\"", token))

		logger.Logger.Println("compiling the copied code")
		output += execCommand(fmt.Sprintf("docker exec ns3-%s sh -c \"cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf\"", token))

		// run the compiled file
		logger.Logger.Println("running the ns3 build")
		execCommand(fmt.Sprintf("docker exec ns3-%s sh -c \"cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf --run file > log.out 2>&1\"", token))

		execCommand(fmt.Sprintf("docker cp ns3-%s:/usr/ns-allinone-3.30.1/ns-3.30.1/log.out %s", token, requestRoot))

		ob, err := ioutil.ReadFile(fmt.Sprintf("%s/log.out", requestRoot))
		if err != nil {
			logger.Logger.Println(fmt.Sprint(err))
			output += "\n>>> Could not capture program output\n"
		}
		output += string(ob)

		// copy pcap to output dir
		logger.Logger.Println("copy output to output directory")
		output += execCommand(fmt.Sprintf("docker exec ns3-%s bash -c \"mkdir -p /output; cp -f /usr/ns-allinone-3.30.1/ns-3.30.1/*.pcap /output\"", token))

		// copy the pcap files back to the host
		logger.Logger.Println("copy output to the host")
		output += execCommand(fmt.Sprintf("docker cp ns3-%s:/output/. %s.", token, requestRoot))

		logger.Logger.Println("Stoping the docker container")
		execCommand(fmt.Sprintf("docker stop ns3-%s", token))

		logger.Logger.Println("removing the stray docker container")
		execCommand(fmt.Sprintf("docker rm ns3-%s", token))

		logger.Logger.Println(fmt.Sprintf("sending response %s", token))
		json.NewEncoder(w).Encode(map[string]string{"token": token, "output": output})
	})

	router.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		buf := new(bytes.Buffer)
		writer := zip.NewWriter(buf)

		_, b, _, _ := runtime.Caller(0)

		// Root folder of this project
		root := filepath.Join(filepath.Dir(b), "../..", fmt.Sprintf("uploads/%s/", token))
		files, _ := ioutil.ReadDir(root)
		for _, file := range files {
			filename := file.Name()

			// Root folder of this project
			filePath := filepath.Join(filepath.Dir(b), "../..", fmt.Sprintf("uploads/%s/%s", token, filename))
			data, err := ioutil.ReadFile(filePath)
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
		WriteTimeout: 600 * time.Second,
		ReadTimeout:  600 * time.Second,
	}

	logger.Logger.Println("Starting server")
	log.Fatal(srv.ListenAndServe())
}
