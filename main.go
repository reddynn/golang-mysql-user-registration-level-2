package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Dbconnect() *sql.DB {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1)/userdb")

	if err != nil {
		panic(err.Error())

	}

	return db
}
func main() {
	http.HandleFunc("/", Homepage)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/signin", Signin)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
	}
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "bad request decoding error", http.StatusBadRequest)
		return
	}
	dbs := Dbconnect()
	var newuser string
	err = dbs.QueryRow("select username from users where username=?", user.Username).Scan(&newuser)
	switch {
	case err == sql.ErrNoRows:
		hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "unable create account", http.StatusInternalServerError)
			return
		}
		_, err = dbs.Query("insert into users(username,password) values(?,?)", user.Username, hashedpassword)
		if err != nil {
			http.Error(w, "server unable to create user", http.StatusInternalServerError)
			return

		}
		w.Write([]byte("user created"))
	case err != nil:
		http.Error(w, " server error, unbale to create account", http.StatusInternalServerError)
		return

	default:
		http.Error(w, "user already existed", http.StatusBadRequest)

	}

}

func Signin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	dbs := Dbconnect()
	var newpassword string
	err = dbs.QueryRow("select password from users where username=?", user.Username).Scan(&newpassword)
	switch {
	case err != nil:
		http.Error(w, "unauthorized no user with this username", http.StatusUnauthorized)
		return

	default:
		err := bcrypt.CompareHashAndPassword([]byte(newpassword), []byte(user.Password))
		if err != nil {
			http.Error(w, "unathorized password wrong", http.StatusUnauthorized)
			break
		}
		w.Write([]byte("welcome" + user.Username))
	}

}
