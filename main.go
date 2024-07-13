package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

var tasks = make(map[int]Task)
var id = 0

func main() {
	http.HandleFunc("/tasks", GetPostMethod)
	http.HandleFunc("/tasks/", idMethod)
	port := ":8080"
	log.Println("Start server")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Error to start server")
	}
	log.Println("Server work")
}

func GetPostMethod(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		viewTask(w)
	case http.MethodPost:
		addTask(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func idMethod(w http.ResponseWriter, r *http.Request){
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method{
	case http.MethodDelete:
		deleteTask(w, id)
	}
}

func viewTask(w http.ResponseWriter) {
	listTasks := make([]Task, 0, len(tasks))
	for _, i := range tasks {
		listTasks = append(listTasks, i)
	}

	json.NewEncoder(w).Encode(listTasks)

}

func addTask(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task.ID = id
	tasks[task.ID] = task
	id++
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func deleteTask(w http.ResponseWriter, id int){
	if _, ok := tasks[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	} 
	delete(tasks, id)
}
