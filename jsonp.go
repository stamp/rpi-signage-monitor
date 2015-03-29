package main

import (
    "github.com/ant0ine/go-json-rest/rest"
    "log"
    "net/http"
	"fmt"
	"time"
	"os/exec"
)

func main() {
    api := rest.NewApi()
    //api.Use(rest.DefaultDevStack...)
    api.Use([]rest.Middleware{
	&rest.TimerMiddleware{},
	&rest.RecorderMiddleware{},
	&rest.PoweredByMiddleware{},
	&rest.RecoverMiddleware{
		EnableResponseStackTrace: true,
	},
	&rest.JsonIndentMiddleware{},
	&rest.ContentTypeCheckerMiddleware{},
    }...)

    api.Use(&rest.JsonpMiddleware{
        CallbackNameKey: "callback",
    })

    pingFromBrowser := make(chan bool);

    api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
	// Try to deliver ping
	select {
		case pingFromBrowser <- true:
		default:
	}
        w.WriteJson(map[string]string{"status": ""})
    }))

    go monitor(pingFromBrowser)

    log.Fatal(http.ListenAndServe(":80", api.MakeHandler()))
}

func monitor(pingFromBrowser chan bool) {
	expire := time.NewTimer(time.Second*60);

	for {
		select {
		case <-expire.C:
			expire.Reset(time.Second*120);

			_, err := exec.Command("/usr/bin/killall","chromium").Output()
			if err != nil {
				fmt.Println("/usr/bin/killall chromium: FAIL - ",err)
				continue
			} 
		case <-pingFromBrowser:
			expire.Reset(time.Second*60);
		}	
	}
}
