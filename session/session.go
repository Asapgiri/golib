package session

import (
	// "dunakeke/config"
	// "dunakeke/dictionary"
	// "dunakeke/logic"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gorilla/sessions"
)

type Config struct {
    Title           string
    SiteTitle       string
    TitleSeparator  string
    MaxImgUploadMB  int64
}

type Auth struct {
    Id          string
    Error       string
    Username    string
    Name        string
    Email       string
    Roles       []string
    IsAdmin     bool
    IsMod       bool
    IsEditor    bool
}

type MetaData map[string]string

type Error struct {
    Type    string
    Message string
}

type ErrorHandler struct {
    errs []Error
}

type Sessioner struct {
    id          string
    Config      Config
    Auth        Auth
    Error       ErrorHandler
    Main        string
    MainDto     any
    Path        string
    Dto         any
    Dictionary  any
    Meta        MetaData
}

var err_list = map[string][]Error{}

func (eh *ErrorHandler) init(id string) {
    eh.errs, ok := err_list[id]
    if !ok {
        err_list[id] = []Error{}
    }
}

func (eh *ErrorHandler) Set(typ string, msg string) {
    append(eh.errs, Error{Type: typ, Message: msg})
}

func (eh *ErrorHandler) Get() []Error {
    errs = eh.errs
    eh.errs = []Error{}
    return errs
}

func generateId() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return base64.StdEncoding.EncodeToString(bytes)[:32]
}


//FIXME: Handle fully separately in every function/session!!
//var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var store = sessions.NewCookieStore([]byte("fsjdglkhdsagjklhads;fjklhasl;kfjs"))
var sessionName = "dunakeke"

func (session *Sessioner) Authenticate(r *http.Request) {
    // TODO: Add request aut header
    real_session, _ := store.Get(r, sessionName)
    uname, _ := real_session.Values["name"].(string)
    id, _ := real_session.Values["id"].(string)

    session.Auth.Username = uname
    session.id = id
    session.Error.init(id)
}

func (session *Sessioner) New(w http.ResponseWriter, r *http.Request, uname string) {
    // FIXME: Store auth headers in database with associated user
    store.MaxAge(86400)
    rsess, _ := store.New(r, sessionName)

    id := generateId()

    rsess.Values["name"] = uname
    rsess.Values["id"] = id
    rsess.Save(r, w)
    session.Auth.Username = uname
    session.id = id
}

func (session *Sessioner) Delete(w http.ResponseWriter, r *http.Request) {
    real_session, _ := store.Get(r, sessionName)
    real_session.Options.MaxAge = -1
    real_session.Save(r, w)
}

func (session *Sessioner) SetError(msg string) {
    session.Auth.Username = ""
    session.Auth.Error = msg
}

func (session *Sessioner) UpdateTitle(config Config, title string) {
    session.Config.Title = config.Title + config.TitleSeparator + title
}

func (md *MetaData) Add(key string, value string) *MetaData {
    (*md)[key] = value
    return md
}
