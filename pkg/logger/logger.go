package logger

import (
	"github.com/regorov/logwriter"
	"log"
	"time"
)

var Logger *log.Logger

func Setup() {
	cfg := &logwriter.Config{
		BufferSize:       0,                  // no buffering
		FreezeInterval:   1 * time.Hour,      // freeze log file every hour
		HotMaxSize:       100 * logwriter.MB, // 100 MB max file size
		CompressColdFile: true,               // compress cold file
		HotPath:          "/var/log/ns3playground",
		ColdPath:         "/var/log/ns3playground/arch",
		Mode:             logwriter.ProductionMode, // write to file only
	}

	lw, err := logwriter.NewLogWriter("ns3playground",
		cfg,
		true, // freeze hot file if exists
		nil)

	if err != nil {
		panic(err)
	}

	Logger = log.New(lw, "mywebserver", log.Ldate|log.Ltime)
	Logger.Println("Logging Setup")

}
