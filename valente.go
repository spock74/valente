package valente

import (
	"errors"
	"log"

	"golang.org/x/net/websocket"
)

const (
	endofmessage = "___ENDOFMESSAGE___"
)

var (
	ProtocolError = errors.New("Protocol Error")
)

//HandlerFunc is a function of handle an event received into websocket.Conn
type HandlerFunc func(*websocket.Conn, *App, []string)

//Form represents the unit of user interaction
type Form interface {
	AddEventHandler(evt string, f HandlerFunc) Form
	Run(ws *websocket.Conn, app *App) error
	Initialize(ws *websocket.Conn) Form
}

//FormImpl its a simple Form
type FormImpl struct {
	trans map[string]HandlerFunc
}

//AddEventHandler add an f function to handle evt event
func (form FormImpl) AddEventHandler(evt string, f HandlerFunc) Form {
	if form.trans == nil {
		form.trans = map[string]HandlerFunc{}
	}
	form.trans[evt] = f
	return form
}

//Run execs the form
func (form FormImpl) Run(ws *websocket.Conn, app *App) error {
	msgs := []string{}
	for {
		msg := ""
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Println("Error on WS Receive", err)
			return err
		}
		if msg == endofmessage {
			break
		} else {
			msgs = append(msgs, msg)
		}
	}
	println("msgs = ", msgs)
	if len(msgs) < 1 {
		return ProtocolError
	}

	f, present := form.trans[msgs[0]]
	if present {
		f(ws, app, msgs)
	} else {
		log.Println("Evt not found", msgs[0])
	}
	return nil
}

//Initialize inits the form
func (form FormImpl) Initialize(ws *websocket.Conn) Form {
	log.Println("FormImpl Initialize")
	return form
}

//App is a Web Application representation
type App struct {
	WS          *websocket.Conn
	Forms       map[string]Form
	Data        map[string]interface{}
	CurrentForm Form
}

//GoTo replace the current form into app
func (app *App) GoTo(formName string) error {
	log.Println("App goto", formName)
	form, present := app.Forms[formName]
	if present {
		app.CurrentForm = form.Initialize(app.WS)
	} else {
		log.Println("[ERROR] Form not registred", formName)
	}
	return nil
}

//Run handle events
func (app *App) Run() {
	app.Data = map[string]interface{}{}
	for {
		err := app.CurrentForm.Run(app.WS, app)
		if err != nil {
			return
		}
	}
}

//Initialize inits the App
func (app *App) Initialize() {
	log.Println("App Initialize")
}

//AddForm add a new form to App
func (app *App) AddForm(name string, f Form) {
	log.Println("AddForm", name, f)
	if app.Forms == nil {
		app.Forms = map[string]Form{}
	}

	app.Forms[name] = f
}
