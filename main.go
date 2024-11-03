package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var tmpl *template.Template
var db *sql.DB

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	var err error
	tmpl, err = template.ParseGlob("./*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

func initDB() {
	dbUser := "root"             // Replace with your MySQL username
	dbPassword := "eDyrr7355608" // Replace with your MySQL password
	dbName := "testdb"           // Replace with your database name
	dbHost := "127.0.0.1"        // Or the host of your MySQL server, if remote

	// Database connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName)

	// Open database connection
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
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

	gRouter.HandleFunc("/gettaskupdateform/{id}", getUpdateTaskForm).Methods("GET")

	gRouter.HandleFunc("/tasks/{id}", updateTask).Methods("PUT", "POST")

	gRouter.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":4000", gRouter))
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "home.html", nil)
}

func fetchTasks(w http.ResponseWriter, r *http.Request) {
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

func getTaskForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "addTaskForm", nil)
}

func getUpdateTaskForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	task, err := getTaskByID(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(w, "updateTaskForm", task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {

	fmt.Println("update func called")
	vars := mux.Vars(r)

	taskItem := r.FormValue("task")

	var taskStatus bool

	fmt.Println(r.FormValue("done"))

	switch strings.ToLower(r.FormValue("done")) {
	case "yes", "on":
		taskStatus = true
	case "no", "off":
		taskStatus = false
	default:
		taskStatus = false
	}

	taskId, _ := strconv.Atoi(vars["id"])

	task := Task{
		taskId, taskItem, taskStatus,
	}

	updateErr := updateTaskById(db, task)

	if updateErr != nil {
		log.Fatal(updateErr)
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todolist", todos)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	err := deleteTaskWithId(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todos", todos)
}

func getTasks(db *sql.DB) ([]Task, error) {
	query := "SELECT id, task, done FROM tasks"

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var todo Task

		rowErr := rows.Scan(&todo.Id, &todo.Task, &todo.Done)

		if rowErr != nil {
			return nil, rowErr
		}

		tasks = append(tasks, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func getTaskByID(db *sql.DB, id int) (*Task, error) {
	query := "SELECT id, task, done FROM tasks WHERE id = ?"

	var task Task

	row := db.QueryRow(query, id)
	err := row.Scan(&task.Id, &task.Task, &task.Done)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no task was found with id %d", id)
		}
		return nil, err
	}

	return &task, nil

}

func updateTaskById(db *sql.DB, task Task) error {
	query := "UPDATE Tasks SET task = ?, done = ? WHERE id = ?"

	result, err := db.Exec(query, task.Task, task.Done, task.Id)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %w", err)
	}

	if rowsAffected == 0 {
		log.Println("no rows updated")
	} else {
		log.Printf("%d row(s) updated\n", rowsAffected)
	}

	return nil
}

func deleteTaskWithId(db *sql.DB, id int) error {
	query := "DELETE FROM tasks WHERE id = ?"

	stmt, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no task found with id%d", id)
	}

	fmt.Printf("deleted %d task(s)\n", rowsAffected)
	return nil
}
