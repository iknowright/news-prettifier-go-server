package app
 
import (
    "net/http"
 
	"github.com/gorilla/mux"
	
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"log"

	helpers "../helpers"
 
	"github.com/gorilla/securecookie"
    "text/template"

    "encoding/json"
)
 
//App struct
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type login struct {
    Owner string
    Template int
    Article_IDs []string
    Current_Article article
}

const (
	host="news-prettifier.cxo5rl1pvafb.us-east-2.rds.amazonaws.com"
	port="5432"
	user="news_manager"
	password="password"
	dbname="news_prettifier"
)

//Initialize method
func (a *App) Initialize() {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

//Run method
func (a *App) Run(addr string) {
	// ListenAndServer needs a port string and Handler which requires ServeHTTP(ResponseWriter, *Request) method
	// The mux.Router implements ServeHTTP(response, *request)
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}
 
func (a *App) InitializeRoutes() {
    // default page

    a.Router.HandleFunc("/{uuid:[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+}", a.HomePageHandler) // GET
    a.Router.HandleFunc("/", a.HomePageHandler) // GET
    
    
    a.Router.HandleFunc("/index/", a.IndexPageHandler) // GET
    a.Router.HandleFunc("/index/{uuid:[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+}", a.IndexPageHandler) // GET
    
    // login

    a.Router.HandleFunc("/login/", a.LoginPageHandler).Methods("GET") // GET
    a.Router.HandleFunc("/login/{uuid:[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+}", a.LoginPageHandler).Methods("GET")
    a.Router.HandleFunc("/login/", a.LoginHandler).Methods("POST")
    a.Router.HandleFunc("/login/{uuid:[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+}", a.LoginHandler).Methods("POST")
 
    // register

    a.Router.HandleFunc("/register", a.RegisterPageHandler).Methods("GET")
    a.Router.HandleFunc("/register", a.RegisterHandler).Methods("POST")
 
    // logout
    a.Router.HandleFunc("/logout/", a.LogoutHandler).Methods("GET")
    a.Router.HandleFunc("/logout/{uuid:[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+-[a-f0-9]+}", a.LogoutHandler).Methods("GET")
    

    // article post
    a.Router.HandleFunc("/article", a.createArticle).Methods("POST")

    // article update settings
    a.Router.HandleFunc("/article_settings", a.updateAccountSettings).Methods("POST")    

    // serve static
    a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
 
    http.Handle("/", a.Router)
    http.ListenAndServe(":8000", nil)
}


var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))
 

// Helper function

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Handlers
 
// for GET
func (a *App) HomePageHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Println("GET home page handler")  
    vars := mux.Vars(request)
    n := article{
        Article_ID: "",
        Title: "",
        Author: "",
        Content: "",
        Origin: "",
    }
    if len(vars) != 0 {
        fmt.Println(vars["uuid"])
        n.Article_ID = vars["uuid"]
        if err := n.getArticle(a.DB); err != nil {
            switch err {
            case sql.ErrNoRows:
                fmt.Println("no such article")
            default:
                fmt.Println("bad query")
            }
        }
    }
    tmpl := template.Must(template.ParseFiles("templates/home.html"))
    tmpl.Execute(response, n)
}

// for GET
func (a *App) LoginPageHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Println("GET login Page handler")
    n := ""
    vars := mux.Vars(request)
    if len(vars) != 0 {
       n = vars["uuid"]
    }
    tmpl := template.Must(template.ParseFiles("templates/login.html"))
    tmpl.Execute(response, n)
}
 
// for POST
func (a *App) LoginHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Println("POST login handler")
    name := request.FormValue("name")
    pass := request.FormValue("password")
    fmt.Println(name)
    fmt.Println(pass)
    redirectTarget := "/"
    n := ""
    vars := mux.Vars(request)
    if len(vars) != 0  {
        n = vars["uuid"]
    }
    if !helpers.IsEmpty(name) && !helpers.IsEmpty(pass) {
        // Database check for user data!
        _userIsValid := a.UserIsValid(name, pass)
 
        if _userIsValid {
            a.SetCookie(name, response)
            redirectTarget = fmt.Sprintf("/index/%s", n)
        } else {
            redirectTarget = "/register"
        }
        fmt.Println(redirectTarget)
    }
    http.Redirect(response, request, redirectTarget, 302)
}
 
// for GET
func (a *App) RegisterPageHandler(response http.ResponseWriter, request *http.Request) {
    var body, _ = helpers.LoadFile("templates/register.html")
    fmt.Fprintf(response, body)
}
 
// for POST
func (a *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
 
    uName := r.FormValue("username")
    email := r.FormValue("email")
    pwd := r.FormValue("password")
	confirmPwd := r.FormValue("confirmPassword")
	
	fmt.Println(uName)
	fmt.Println(email)
	fmt.Println(pwd)
	fmt.Println(confirmPwd)
	
 
    _uName, _email, _pwd, _confirmPwd := false, false, false, false
    _uName = !helpers.IsEmpty(uName)
    _email = !helpers.IsEmpty(email)
    _pwd = !helpers.IsEmpty(pwd)
    _confirmPwd = !helpers.IsEmpty(confirmPwd)
 
    if _uName && _email && _pwd && _confirmPwd {
        fmt.Fprintln(w, "Username for Register : ", uName)    
        fmt.Fprintln(w, "Email for Register : ", email)     
        fmt.Fprintln(w, "Password for Register : ", pwd)  
        fmt.Fprintln(w, "ConfirmPassword for Register : ", confirmPwd)    
        if pwd == confirmPwd {
            n := account{ Username:uName, Password:pwd, Email:email}
            if err := n.createAccount(a.DB); err != nil {
                fmt.Println("I'm here with error!")                            
                return
            }         
        } else {
            fmt.Fprintln(w, "Password not match!")
        }
    } else {
        fmt.Fprintln(w, "This fields can not be blank!")
    }
}

type login_data struct {
    Owner string
    Size int
    Color int
    Articles []article
    Current_Article article
}

// for GET
func (a *App) IndexPageHandler(response http.ResponseWriter, request *http.Request) {
    vars := mux.Vars(request)
    redirectTarget := "/"
    if len(vars) != 0 { 
        redirectTarget = fmt.Sprintf("/%s", vars["uuid"])        
    }
    userName := a.GetUserName(request)
    if !helpers.IsEmpty(userName) {
        curr_article := article{
            Article_ID: "",
            Title: "",
            Author: "",
            Content: "",
            Origin: "",
        }
        fmt.Println("im in login handler")    
        if len(vars) != 0 {
            curr_article.Article_ID = vars["uuid"]
            if err := curr_article.getArticle(a.DB); err != nil {
                switch err {
                case sql.ErrNoRows:
                    fmt.Println("no such article")
                default:
                    fmt.Println("bad query")
                }
            }
        }
        user := account{ Username:userName }
        if err := user.getAccount(a.DB); err != nil {
            switch err {
            case sql.ErrNoRows:
                fmt.Println("no this user");
            default:
                fmt.Println("no this user");
            }
        }
        user_articles, err := getArticles(a.DB, userName)
        if err != nil {
            fmt.Println("error in get multiple article")
            return
        }
        history_used := false
        for _, elem := range user_articles {
            if elem.Title == curr_article.Title {
                history_used = true
                break;
            }
        }
        if !history_used {
            if err := curr_article.updateArticleUser(a.DB, userName); err != nil {
                switch err {
                case sql.ErrNoRows:
                    fmt.Println("no such article")
                default:
                    fmt.Println("bad query")
                }
            }
        }
        loginData := login_data{
            Owner: userName,
            Size: user.Size,
            Color: user.Color,
            Articles: user_articles,
            Current_Article: curr_article,
        }
        fmt.Println(loginData)
        tmpl := template.Must(template.ParseFiles("templates/index.html"))
        tmpl.Execute(response, loginData)
    } else {
        http.Redirect(response, request, redirectTarget, 302)
    }
}
 
// for POST
func (a *App) LogoutHandler(response http.ResponseWriter, request *http.Request) {
    vars := mux.Vars(request)
    redirectTarget := "/"
    if len(vars) != 0 { 
        redirectTarget = fmt.Sprintf("/%s", vars["uuid"])        
    }
    a.ClearCookie(response)
    http.Redirect(response, request, redirectTarget, 302)
}
 
// Cookie
 
func (a *App) SetCookie(userName string, response http.ResponseWriter) {
    value := map[string]string{
        "name": userName,
    }
    if encoded, err := cookieHandler.Encode("cookie", value); err == nil {
        cookie := &http.Cookie{
            Name:  "cookie",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
    }
}
 
func (a *App) ClearCookie(response http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:   "cookie",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(response, cookie)
}
 
func (a *App) GetUserName(request *http.Request) (userName string) {
    if cookie, err := request.Cookie("cookie"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("cookie", cookie.Value, &cookieValue); err == nil {
            userName = cookieValue["name"]
        }
    }
    return userName
}

// Check User Exists

func (a *App) UserIsValid(uName, pwd string) bool {
    // DB simulation

    n := account{Username: uName}
    fmt.Println(uName)
	if err := n.getAccount(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			fmt.Println("no this user");
		default:
			fmt.Println("no this user");
		}
		return false
    }
	fmt.Printf("%+v\n", n);
    _isValid := false
 
    if uName == n.Username && pwd == n.Password {
        _isValid = true
    } else {
        _isValid = false
    }
    fmt.Println(_isValid)
    return _isValid
}

// Handling Article

func (a *App) createArticle(w http.ResponseWriter, r *http.Request) {
	var n article
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
    defer r.Body.Close()
    
    fmt.Println(n)

	if err := n.createArticle(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, n)
}

func (a *App) updateAccountSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST update account setting handler")
    var n account
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
    defer r.Body.Close()
    
    fmt.Println(n)
	if err := n.updateAccountSettings(a.DB); err != nil {
		fmt.Println("Can't update size and color")
		return
	}

}