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

type NoticeType struct {
    INFO    string
    SUCCESS string
    WARNING string
    DANGER  string
}

var NOTICE = NoticeType{
    INFO:       "info",
    SUCCESS:    "success",
    WARNING:    "warning",
    DANGER:     "danger",
}

type Notice struct {
    Type    string
    Message string
}

type NoticeHandler struct {
    id string
    Notices []Notice
}

type Sessioner struct {
    id          string
    Config      Config
    Auth        Auth
    Notice      NoticeHandler
    Main        string
    MainDto     any
    Path        string
    Dto         any
    Dictionary  any
    Meta        MetaData
}

var err_list = map[string][]Notice{}

func (eh *NoticeHandler) init(id string) {
    errs, ok := err_list[id]
    if !ok {
        errs = []Notice{}
        err_list[id] = errs
    }
    eh.id = id
    eh.Notices = errs
}

func (eh *NoticeHandler) Set(typ string, msg string) {
    eh.Notices = append(eh.Notices, Notice{Type: typ, Message: msg})
    err_list[eh.id] = eh.Notices
}

func (eh *NoticeHandler) Clean() {
    delete(err_list, eh.id)
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

func (session *Sessioner) Authenticate(w http.ResponseWriter, r *http.Request) {
    // TODO: Add request aut header
    real_session, _ := store.Get(r, sessionName)
    uname, _ := real_session.Values["name"].(string)
    id, _ := real_session.Values["id"].(string)

    if "" == id {
        id = generateId()
        real_session.Values["id"] = id
        real_session.Save(r, w)
    }

    session.Auth.Username = uname
    session.id = id
    session.Notice.init(id)
}

func (session *Sessioner) New(w http.ResponseWriter, r *http.Request, uname string) {
    // FIXME: Store auth headers in database with associated user
    store.MaxAge(3000000)
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
