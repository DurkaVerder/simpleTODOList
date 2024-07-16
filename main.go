package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Task struct {
	ID      int    `json:"id"`
	Header  string `json:"header"`
	Content string `json:"content"`
}

var db *sql.DB

func main() {
	runPostgreSQL()
	defer db.Close()
	startServer()
}

func runPostgreSQL() {
	connectStr := "user=postgres password=durka dbname=store sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting database: ", err)
	}
}

func startServer() {
	http.HandleFunc("/tasks", handleRequest)
	http.HandleFunc("/tasks/", handleRequestWithAddingParameter)
	port := ":8080"
	log.Println("Start server")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Error to start server")
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTask(w)
	case http.MethodPost:
		postTask(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func handleRequestWithAddingParameter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		log.Fatal("Error convert string to int: ", err)
	}
	switch r.Method {
	case http.MethodPut:
		putTask(w, r, id)
	case http.MethodDelete:
		deleteTask(w, id)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getTask(w http.ResponseWriter) {
	selectStr := `SELECT id, header, content FROM task`
	rows, err := db.Query(selectStr)
	if err != nil {
		log.Fatal("Error select-request: ", err)
	}
	defer rows.Close()
	tasks := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Header, &t.Content); err != nil {
			log.Fatal("Error scan rows: ", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Fatal("Error rows: ", err)
	}
	json.NewEncoder(w).Encode(tasks)
	log.Println("Complete GET-request")
}

func postTask(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var t Task
	if err := json.Unmarshal(body, &t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	postSQL := `INSERT INTO task (header, content) VALUE ($1, $2)`
	if _, err := db.Exec(postSQL, t.Header, t.Content); err != nil {
		log.Fatal("Error insert data in database: ", err)
	}
	json.NewEncoder(w).Encode(t)
	log.Println("Complete POST-request")
}

func deleteTask(w http.ResponseWriter, id int) {
	deleteSQL := `DELETE FROM task WHERE id = $1`
	if _, err := db.Exec(deleteSQL, id); err != nil {
		log.Fatal("Error delete: ", err)
	}
	log.Println("Complete DELETE-request")
}

func putTask(w http.ResponseWriter, r *http.Request, id int) {
	PutSQL := `UPDATE task SET content = $1 WHERE id = $2`

	newContent, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var t Task
	if err := json.Unmarshal(newContent, &t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := db.Exec(PutSQL, t.Content, id); err != nil {
		log.Fatal("Error change: ", err)
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Complete PUT-request")
}
