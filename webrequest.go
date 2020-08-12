package webrequest

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"time"

	//"madman/servicemanager/web/dashboard"
	"github.com/madman22/dashboard"
)

//TODO implement more route options

func init() {
	fmt.Println("INITING GOB REQUEST TYPES")
	gob.Register(Guest)
	gob.Register(&WebUser{})
	gob.Register(&WebRoute{})
	gob.Register(time.Time{})
	gob.Register(&WebRequest{})
}

//TODO use Template in web server instead of element,string,error, just element,error and user settings/access determines template
type WebRequest struct {
	ID string
	WebUser
	WebRoute
	Timestamp time.Time
	//Request   *http.Request
	Form       url.Values
	RemoteAddr string
	Template   string
	Uri        string
}

func (wr *WebRequest) String() string {
	//var out []string
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
	wr.Password = ""
	wr.AccessLevel = Guest
	wr.RemoteAddr = ""
	wr.Uri = ""
	wr.Template = ""
	wr.Form = make(url.Values)
	wr.Timestamp = time.Time{}
}

type WebRoute struct {
	Service string
	Section string
	Action  string
	Item    string
	//Options []string
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
	//var out string
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

type WebUser struct {
	ID       string
	Username string
	Password string
	AccessLevel
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

type WebFunc func(*WebRequest) (dashboard.Element, string, error)

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
		return errors.New("Key now found")
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

func (wm WebMap) Do(wr *WebRequest) (dashboard.Element, string, error) {
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
	return nil, "", errors.New("Route not found" + wm.String())
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
			body += ` _____ `
		}
		body += route.String()
	}
	return body
}

//TODO create structure for routes that are interchangable, and can be used to auto create the menu
/*



type WebRoute2 struct {
	Path RouteList
}

type RouteList struct {

}

type RouteQualifier interface {
	HREF() string
	Value() string
	Next() RouteQualifier
}

/*

func (s *service) Route(RouteQualifier) (Route, error)


*/
