package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var tmpl *templ.Template
var db *sql.DB

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl, _ = template.ParseGlob("templates/*.html")
}

func initDB() {
	var err error

	db, err = sql.Open("mysql", "root:root@(localhost:3333)/testdb?parseTime=true")

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	gRouter := mux.NewRouter()

	initDB()
	defer db.Close()

	gRouter.HandleFunc("/", Homepage)

	gRouter.HandleFunc("/tasks", fetchTasks).Methods("GET")

	gRouter.HandleFunc("/newtaskform", getTaskForm)

	gRouter.HandleFunc("/tasks", addTask).Methods("POST")

	gRouter.HanldeFunc("/gettaskupdateform/{id}", getUpdateTaskFrom).Methods("GET")

	gRouter.HandleFunc("/tasks/{id}", updateTask).Methods("PUT", "POST")

	gRouter.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	http.ListenAndServe(":4000", gRouter)
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "home.html", nil)
}

func fetchTasks(w http.ResponseWriter, r *http.Response) {
	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todolist", todos)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	task := r.FormValue("task")

	fmt.Println(task)

	query := "INSERT INTO Tasks (task, done) VALUES (?, ?)"
	stmt, err := db.Prepare(query)

	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	_, executeErr := stmt.Exec(task, 0)

	if executeErr != nil {
		log.Fatal(executeErr)
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todolist", todos)
}

func gettaskupdateform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	task, err := getTaskByID(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(w, "updateTaskForm", task)
}
