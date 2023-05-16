package updater

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type SaltResult struct {
	Salt string
	Sign string
	Time int64
}

type UpdateReq struct {
	Salt          string
	Time          int64
	Sign          string
	User          string
	Pass          string
	Domain        string
	Address       string
	ClientAddress string
	Code          string
}

type Handler interface {
	Update(domain string, address string) error
}

type Updater struct {
	ServerKey string
	Username  string
	Password  string
	Handlers  map[string]Handler
	Domains   map[string]string
	Aliases   map[string][]string
}

func (u *Updater) RegisterHandler(name string, handler Handler) {
	if u.Handlers == nil {
		u.Handlers = make(map[string]Handler)
	}
	u.Handlers[name] = handler
}

func (u *Updater) RegisterDomain(domain string, handler string) {
	if u.Domains == nil {
		u.Domains = make(map[string]string)
	}
	u.Domains[domain] = handler
}

func (u *Updater) GenerateSalt() *SaltResult {
	salt := randSalt(10)
	now := time.Now().Unix()
	sign := u.sign(salt, now)

	return &SaltResult{
		Salt: salt,
		Sign: sign,
		Time: now,
	}
}

func (u *Updater) Update(r *UpdateReq) (interface{}, error) {
	if time.Unix(r.Time, 0).Add(10 * time.Second).Before(time.Now()) {
		return nil, errors.New("Salt value too old")
	}

	if u.sign(r.Salt, r.Time) != r.Sign {
		return nil, errors.New("Invalid signature")
	}

	if !u.auth(r.Pass, r.Salt) {
		return nil, errors.New("Invalid login attempt")
	}

	switch r.Code {
	case "0":
		return u.updateDomain(r.Domain, r.Address)
	case "1":
		return u.updateDomain(r.Domain, "")
	case "2":
		return u.updateDomain(r.Domain, r.ClientAddress)
	default:
		return nil, errors.New("Invalid client request code")
	}
}

func (u *Updater) updateDomain(domain string, address string) (string, error) {
	handlerName, ok := u.Domains[domain]
	if !ok {
		return "", errors.New("Invalid domain")
	}

	handler, ok := u.Handlers[handlerName]
	if !ok {
		return "", errors.New("Invalid domain handler")
	}

	err := handler.Update(domain, address)
	if err != nil {
		return "", err
	}

	aliases, ok := u.Aliases[domain]
	if ok {
		for _, alias := range aliases {
			_, err := u.updateDomain(alias, address)
			if err != nil {
				return "", err
			}
		}
	}

	return address, nil
}

func (u *Updater) sign(salt string, ts int64) string {
	val := fmt.Sprintf("%s.%d.%s", salt, ts, u.ServerKey)
	return md5hex(val)
}

func (u *Updater) auth(pass, salt string) bool {
	val := fmt.Sprintf("%s.%s", md5hex(u.Password), salt)
	if md5hex(val) != pass {
		return false
	}
	return true
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSalt(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func md5hex(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
