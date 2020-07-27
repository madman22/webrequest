package webrequest

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WebRequest struct {
	ID string
	WebUser
	WebRoute
	RemoteAddr   string
	TemplateName string
	Uri          string
	Timestamp    time.Time
	Form         url.Values
}

type WebUser struct {
	ID       string
	Username string
	AccessLevel
}

type WebRoute struct {
	Service string
	Section string
	Action  string
	Item    string
}

func (wr *WebRequest) String() string {
	out := []string{"WebRequest:", wr.Timestamp.Format(time.Stamp)}
	if len(wr.ID) > 0 {
		out = append(out, "ID:"+wr.ID)
	}
	route := wr.WebRoute.String()
	if len(route) > 0 {
		out = append(out, route)
	}
	out = append(out, wr.WebUser.String())
	if len(wr.RemoteAddr) > 0 {
		out = append(out, "Addr:"+wr.RemoteAddr)
	}

	return strings.Join(out, " ")
}

func (wr *WebRequest) Reset(id string) {
	wr.ID = id
	wr.Service = ""
	wr.Section = ""
	wr.Action = ""
	wr.Item = ""
	wr.WebUser.ID = ""
	wr.Username = ""
	wr.AccessLevel = Guest
	wr.RemoteAddr = ""
	wr.Form = make(url.Values)
	wr.Timestamp = time.Now()
}

func (wr *WebRoute) String() string {
	out := []string{"WebRoute:"}
	if len(wr.Service) > 0 {
		out = append(out, "Service:"+wr.Service)
	}
	if len(wr.Section) > 0 {
		out = append(out, "Section:"+wr.Section)
	}
	if len(wr.Action) > 0 {
		out = append(out, "Action:"+wr.Action)
	}
	if len(wr.Item) > 0 {
		out = append(out, "Item:"+wr.Item)
	}
	if len(out) == 1 {
		return ""
	}
	return strings.Join(out, " ")
}

func (wr *WebRoute) HREF() string {
	out := "/"
	if len(wr.Service) < 1 {
		return out
	}
	out += wr.Service
	if len(wr.Section) < 1 {
		return out
	}
	out += "/" + wr.Section
	if len(wr.Action) < 1 {
		return out
	}
	out += "/" + wr.Action
	if len(wr.Item) < 1 {
		return out
	}
	out += "/" + wr.Item
	return out
}

func (wu *WebUser) String() string {
	out := []string{"User:"}
	if len(wu.ID) > 0 {
		out = append(out, "ID:"+wu.ID)
	}
	if len(wu.Username) > 0 {
		out = append(out, "Username:"+wu.Username)
	}
	out = append(out, "Access:"+wu.AccessLevel.String())
	return strings.Join(out, " ")
}

type AccessLevel int

func (a AccessLevel) String() string {
	switch a {
	case Guest:
		return "Guest"
	case User:
		return "User"
	case Customer:
		return "Customer"
	case Tech:
		return "Technician"
	case Admin:
		return "Administrator"
	case System:
		return "System"
	default:
		return strconv.Itoa(int(a))
	}
}

func (a AccessLevel) CheckAccess(newa AccessLevel) bool {
	return a >= newa
}

const (
	Guest AccessLevel = iota
	User
	Customer
	Tech
	Admin
	System
)

func ParseAccessLevel(input string) (AccessLevel, error) {
	input = strings.ToLower(input)
	switch input {
	case "guest":
		return Guest, nil
	case "user":
		return User, nil
	case "customer":
		return Customer, nil
	case "tech":
		return Tech, nil
	case "admin":
		return Admin, nil
	case "system":
		return System, nil
	default:
		return Guest, errors.New("Cannot parse input to access level:" + input)
	}
}

type TempalteName string

type WebFunc func(*WebRequest) (WebElement, TempalteName, error)

type WebWriter interface {
	Write([]byte) (int, error)
}

type WebElement interface {
	HTML(*WebRequest) ([]byte, error)
	Write(WebWriter, *WebRequest) (int, error)
	WriteAndReplace(WebWriter, *WebRequest, ...WebElement) (int, error)
	WriteWithChildren(WebWriter, *WebRequest, ...WebElement) (int, error)
}

type WebMap map[WebRoute]WebFunc

func (wm WebMap) Add(service, section, action, item string, f WebFunc) error {
	wm[WebRoute{Service: service, Section: section, Action: action, Item: item}] = f
	return nil
}

func (wm WebMap) Remove(service, section, action, item string) error {
	key := WebRoute{Service: service, Section: section, Action: action, Item: item}
	if _, ok := wm[key]; ok {
		delete(wm, key)
	} else {
		return errors.New("Key not found")
	}
	return nil
}

func (wm WebMap) RemoveService(service string) (int, error) {
	count := 0
	for id, _ := range wm {
		if id.Service == service {
			delete(wm, id)
			count++
		}
	}
	return count, nil
}

func (wm WebMap) Do(wr *WebRequest) (WebElement, TempalteName, error) {
	route := wr.WebRoute
	if len(wm) < 1 {
		return nil, "", errors.New("No Routes Registered")
	}
	if f, ok := wm[route]; ok {
		return f(wr)
	}
	if len(route.Item) > 0 {
		route.Item = ""
		if f, ok := wm[route]; ok {
			return f(wr)
		}
	}
	if len(route.Action) > 0 {
		route.Action = ""
		if f, ok := wm[route]; ok {
			return f(wr)
		}
	}
	if len(route.Section) > 0 {
		route.Section = ""
		if f, ok := wm[route]; ok {
			return f(wr)
		}
	}
	return nil, "", errors.New("Route not found" + wr.WebRoute.String())
}

func (wm WebMap) Merge(nwm WebMap, overwrite bool) error {
	if wm == nil {
		wm = make(WebMap)
	}
	if nwm == nil {
		return nil
	}
	if len(nwm) < 1 {
		return nil
	}
	for route, f := range nwm {
		if _, ok := wm[route]; ok && !overwrite {
			return errors.New("Route already exists:" + route.String())
		}
		wm[route] = f
	}
	return nil
}

func (wm WebMap) String() string {
	var body string
	for route, _ := range wm {
		if len(body) > 0 {
			body += `, `
		}
		body += route.String()
	}
	return body
}
