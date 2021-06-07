package plugin

import (
	"encoding/json"
	"net/http"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/pkg/jwt"
	"sync"
	"time"
)

type Login struct {
	Name string `json:"name"`
	RequestType string `json:"request_type"`
}

type Token struct {
	JWT string `json:"jwt"`
}

type Event struct {
	Message string `json:"message"`
}

type Video struct {
	Message string `json:"message"`
	Link    string `json:"link"`
}

var EventObservers = map[string]chan *Event{}
var EventCreateMutex sync.Mutex
var EventRemoveMutex sync.Mutex
var EventNotifyMutex sync.Mutex

var IsIncomingCall = false
var AnsweredPlugin = ""
var AnswerMutex sync.Mutex
var CancelMutex sync.Mutex

var IntercomObserver chan *Event
var IsIntercomObserverOpen = false
var IntercomMessage = ""

func RegisterPlugin(w http.ResponseWriter, r *http.Request) {
	login := &Login{}

	err := json.NewDecoder(r.Body).Decode(login)

	if err != nil {
		http.Error(w, "invalid body", http.StatusForbidden)
		return
	}

	tokenString, err := jwt.GenerateTokenForPlugin(login.Name)

	if err != nil {
		http.Error(w, "generate error", http.StatusForbidden)
		return
	}

	token := &Token{
		JWT: tokenString,
	}

	err = json.NewEncoder(w).Encode(token)

	if err != nil {
		http.Error(w, "encode error", http.StatusForbidden)
		return
	}
}

func IncomingCall(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	video := &Video{}

	err := json.NewDecoder(r.Body).Decode(video)

	if err != nil {
		http.Error(w, "invalid body", http.StatusForbidden)
		return
	}

	IsIncomingCall = true
	AnsweredPlugin = ""

	result := &Event{
		Message: "incoming",
	}

	EventNotifyMutex.Lock()

	for _, value := range EventObservers {
		value <- result
	}

	EventNotifyMutex.Unlock()
}

func RejectedCall(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	IsIncomingCall = false
	AnsweredPlugin = ""
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	if _, ok := EventObservers[id]; ok {
		http.Error(w, "already subscribed", http.StatusForbidden)
		return
	}

	if IsIncomingCall && AnsweredPlugin != "" {
		result := &Event{
			Message: "incoming",
		}

		err := json.NewEncoder(w).Encode(result)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		return
	}

	event := make(chan *Event, 1)

	EventCreateMutex.Lock()
	EventObservers[id] = event
	EventCreateMutex.Unlock()

	select {
	case result := <-event:
		err := json.NewEncoder(w).Encode(result)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	case <-time.After(time.Second * 60):
		timeoutEvent := &Event{
			Message: "",
		}

		err := json.NewEncoder(w).Encode(timeoutEvent)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		<-event
	}

	EventRemoveMutex.Lock()
	delete(EventObservers, id)
	EventRemoveMutex.Unlock()

	close(event)
}

func Answer(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	AnswerMutex.Lock()

	if AnsweredPlugin == "" && IsIncomingCall {
		AnsweredPlugin = id

		if IsIntercomObserverOpen {
			answer := &Event{
				Message: "answer",
			}

			IntercomObserver <- answer
		} else {
			IntercomMessage = "answer"
		}

		message := &Video{
			Message: "answered",
			Link: "link",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	} else if AnsweredPlugin != "" && IsIncomingCall {
		message := &Video{
			Message: "busy",
			Link: "",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	} else {
		message := &Video{
			Message: "incoming false",
			Link: "",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	}

	AnswerMutex.Unlock()
}

func Cancel(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	CancelMutex.Lock()

	if AnsweredPlugin == id && IsIncomingCall {
		AnsweredPlugin = id

		if IsIntercomObserverOpen {
			answer := &Event{
				Message: "cancel",
			}

			IntercomObserver <- answer
		} else {
			IntercomMessage = "cancel"
		}

		message := &Event{
			Message: "canceled",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	} else if AnsweredPlugin != id && IsIncomingCall {
		message := &Event{
			Message: "busy",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	} else {
		message := &Event{
			Message: "incoming false",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	}

	CancelMutex.Unlock()
}

func IntercomCommand(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" || IsIntercomObserverOpen {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	if _, ok := EventObservers[id]; ok {
		http.Error(w, "already subscribed", http.StatusForbidden)
		return
	}

	if IntercomMessage != "" {
		result := &Event{
			Message: IntercomMessage,
		}

		err := json.NewEncoder(w).Encode(result)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		return
	}

	IntercomObserver = make(chan *Event, 1)
	IsIntercomObserverOpen = true
	IntercomMessage = ""

	select {
	case result := <-IntercomObserver:
		err := json.NewEncoder(w).Encode(result)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	case <-time.After(time.Second * 60):
		timeoutEvent := &Event{
			Message: "",
		}

		err := json.NewEncoder(w).Encode(timeoutEvent)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		<-IntercomObserver
	}

	IsIntercomObserverOpen = false
	close(IntercomObserver)
}

func Open(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	if !IsIncomingCall {
		message := &Event{
			Message: "rejected",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		return
	}

	if AnsweredPlugin == id {
		if IsIntercomObserverOpen {
			answer := &Event{
				Message: "open",
			}

			IntercomObserver <- answer
		} else {
			IntercomMessage = "open"
		}

		message := &Event{
			Message: "opened",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		AnsweredPlugin = ""
	} else {
		message := &Event{
			Message: "wrong id",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	}
}

func Reject(w http.ResponseWriter, r *http.Request) {
	id := auth.GetLoginPluginState(r.Context())

	if id == "" {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	if !IsIncomingCall {
		message := &Event{
			Message: "rejected",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		return
	}

	if AnsweredPlugin == id {
		if IsIntercomObserverOpen {
			answer := &Event{
				Message: "reject",
			}

			IntercomObserver <- answer
		} else {
			IntercomMessage = "reject"
		}

		message := &Event{
			Message: "rejected",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}

		AnsweredPlugin = ""
		IsIncomingCall = false
	} else {
		message := &Event{
			Message: "wrong id",
		}

		err := json.NewEncoder(w).Encode(message)

		if err != nil {
			http.Error(w, "encode error", http.StatusForbidden)
			return
		}
	}
}
