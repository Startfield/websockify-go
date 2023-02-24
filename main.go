package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Startfield/websockify-go/websockify"
)

type appConfig struct {
	targetAdd string
	runOnce   bool
	webServer bool
}

var fileHandler http.Handler
var config appConfig = appConfig{}

var (
	logger        *log.Logger
	verboseLogger *log.Logger
)

func main() {
	helpFalg := flag.Bool("h", false, "Print Help")
	verboseFlag := flag.Bool("v", false, "Verbose")
	cert := flag.String("cert", "", "SSL certificate file")
	key := flag.String("key", "", "SSL key file")
	webdir := flag.String("web", "", "Serve files from DIR.")
	runOnceFlag := flag.Bool("run-once", false, "handle a single WebSocket connection and exit")
	flag.Parse()
	if *helpFalg {
		flag.PrintDefaults()
		return
	}
	logger = log.New(os.Stdout, "\n", log.Ldate|log.Ltime)
	if *verboseFlag {
		verboseLogger = log.New(os.Stdout, "\n", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		verboseLogger = log.New(ioutil.Discard, "", log.Ldate)
	}
	config.runOnce = *runOnceFlag

	listenadd := flag.Arg(0)
	config.targetAdd = flag.Arg(1)
	ssllog := " - No SSL/TLS support (no cert file)\n"
	if len(*cert) > 0 {
		ssllog = " - SSL/TLS support\n"
	}
	logger.Printf("WebSocket server settings:\n"+
		" - Listen on %s\n"+
		ssllog+
		" - proxying %s\n", listenadd, config.targetAdd)
	//http.Handle("/z/", http.FileServer(http.Dir("./")))
	if len(*webdir) > 0 {
		config.webServer = true
		fileHandler = http.FileServer(http.Dir(*webdir))
	}
	wsock := websockify.Websockify{Target: config.targetAdd}
	http.HandleFunc("/", wsock.WSNoErr)
	if len(*cert) > 0 {
		if err := http.ListenAndServeTLS(listenadd, *cert, *key, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(listenadd, nil); err != nil {
			log.Fatal(err)
		}
	}
}
