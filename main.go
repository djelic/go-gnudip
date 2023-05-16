package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"jeli.cc/go-gnudip/cloudflare"
	"jeli.cc/go-gnudip/server"
	"jeli.cc/go-gnudip/updater"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var baseTemplate = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
                      "http://www.w3.org/TR/html4/loose.dtd">
<html>
<head>
<title>
GnuDIP Update Server
</title>
{{range $key, $val := .Meta}}
<meta name="{{ $key }}" content="{{ $val }}">
{{end}}
</head>
<body>
<center>
<h2>GnuDIP Update Server</h2>
{{ .Msg }}
</center>
</body>
</html>
`

func main() {
	rand.Seed(time.Now().UnixNano())

	c, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.New("base").Parse(baseTemplate))

	s := &server.Server{
		Updater: &updater.Updater{
			ServerKey: c.ServerKey,
			Username:  c.Username,
			Password:  c.Password,
			Domains:   c.Domains,
			Aliases:   c.Aliases,
			Handlers:  map[string]updater.Handler{},
		},
		Template: tmpl,
	}

	if c.CfApiToken != "" {
		cfa, _ := cf.NewWithAPIToken(c.CfApiToken)
		s.Updater.Handlers["cloudflare"] = &cloudflare.Handler{API: cfa}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/gnudip/cgi-bin/gdipupdt.cgi", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery == "" {
			s.GenerateSalt(w, r)
			return
		}

		s.Update(w, r)
	})

	log.Println("listening on " + c.Addr)
	http.ListenAndServe(c.Addr, r)
}
