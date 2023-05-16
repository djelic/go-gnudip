package server

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"jeli.cc/go-gnudip/updater"
)

type Server struct {
	Updater  *updater.Updater
	Template *template.Template
}

type templateParams struct {
	Meta map[string]interface{}
	Msg  string
}

func (s *Server) GenerateSalt(w http.ResponseWriter, r *http.Request) {
	sr := s.Updater.GenerateSalt()

	err := s.Template.Execute(w, templateParams{
		Msg: "Salt generated",
		Meta: map[string]interface{}{
			"salt": sr.Salt,
			"time": sr.Time,
			"sign": sr.Sign,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (s *Server) Update(w http.ResponseWriter, r *http.Request) {
	req, err := updateReq(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	req.ClientAddress = strings.Split(r.RemoteAddr, ":")[0]
	if req.ClientAddress == "" {
		req.ClientAddress = "0.0.0.0"
	}

	addr, err := s.Updater.Update(req)

	msg := "Successful update request"
	if req.Code == "1" {
		msg = "Successful offline request"
	}
	meta := map[string]interface{}{"retc": 0}

	if err != nil {
		msg = "Failed update request: " + err.Error()
		if req.Code == "1" {
			msg = "Failed offline request: " + err.Error()
		}
		meta["retc"] = 1
	} else {
		if req.Code == "2" {
			meta["addr"] = addr
		}
	}

	s.Template.Execute(w, templateParams{
		Meta: meta,
		Msg:  msg,
	})
}

func updateReq(v url.Values) (*updater.UpdateReq, error) {
	var r updater.UpdateReq

	// the "salt" from the first response ("salt=")
	r.Salt = v.Get("salt")
	if r.Salt == "" {
		return nil, errors.New("`salt` is missing")
	}

	// the "time salt generated" value from the first response ("time=")
	time0 := v.Get("time")
	if time0 == "" {
		return nil, errors.New("`time` is missing")
	}
	ts, err := strconv.ParseInt(time0, 10, 64)
	if err != nil {
		return nil, err
	}
	r.Time = ts

	// the "signature" from the first response ("sign=")
	r.Sign = v.Get("sign")
	if r.Sign == "" {
		return nil, errors.New("`sign` is missing")
	}

	// the GnuDIP user name ("user=")
	r.User = v.Get("user")
	if r.User == "" {
		return nil, errors.New("`user` is missing")
	}

	// the MD5 digested password created above ("pass=")
	r.Pass = v.Get("pass")
	if r.Pass == "" {
		return nil, errors.New("`pass` is missing")
	}

	// the GnuDIP domain name ("domn=")
	r.Domain = v.Get("domn")
	if r.Domain == "" {
		return nil, errors.New("`domn` is missing")
	}

	// the IP address to be registered, if the request code is "0" ("addr=")
	r.Address = v.Get("addr")
	if r.Address == "" {
		return nil, errors.New("`addr` is missing")
	}

	// the server "request code" ("reqc="):
	//     "0" - register the address passed with this request
	//     "1" - go offline
	//     "2" - register the address you see me at, and pass it back to me
	r.Code = v.Get("reqc")
	if r.Code == "" {
		return nil, errors.New("`reqc` is missing")
	}

	return &r, nil
}
