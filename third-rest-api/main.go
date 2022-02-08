package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
	"third-rest-api/taskstore"
)

type taskServer struct { // Структура taskServer с полем store с типом TaskStore пакета taskstore
	store *taskstore.TaskStore
}

func NewTaskServer() *taskServer {
	store := taskstore.New() // Обяевляем стор, в котором есть мьютекс, айди записи и мапа с нашими данными
	return &taskServer{store: store} // Возвращаем taskServer со store равному store
}

func main() {
	router := mux.NewRouter() // NewRouter является экземпляром маршрутизатора.
	router.StrictSlash(true) // TRUE: Когда путь "/ путь /", то он будет перенаправлен на "/ Path /" при доступе "/ путь".
							 // FALSE: Когда путь «/ путь», он не будет соответствовать этому пути при доступе «/ пути /»
	server := NewTaskServer() // Присваиваем конструктор нашего сервера

	router.HandleFunc("/task/", server.createTaskHandler).Methods("POST")
	router.HandleFunc("/task/", server.getAllTasksHandler).Methods("GET")
	router.HandleFunc("/task/", server.deleteAllTasksHandler).Methods("DELETE")
	router.HandleFunc("/task/{id:[0-9]+}/", server.getTaskHandler).Methods("GET") // Используется регулярное выражение
	// любой символ от 0 до 9
	router.HandleFunc("/task/{id:[0-9]+}/", server.deleteTaskHandler).Methods("DELETE")
	router.HandleFunc("/tag/{tag}/", server.tagHandler).Methods("GET")
	router.HandleFunc("/due/{year:[0-9]+}/{month:[0-9]+}/{day:[0-9]+}/", server.dueHandler).Methods("GET")

	log.Fatal(http.ListenAndServe("localhost:8080", router)) // Запускаем сервер на 8080 порту
}

func (ts *taskServer) createTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling task create at %s\n", req.URL.Path)

	type RequestTask struct { // Создаем структуру с колекцией из трёх полей соответствующего типа
		Text string `json:"text"`
		Tags []string `json:"tags"`
		Due time.Time `json:"due"`
	}

	type ResponseId struct { // Создаем структуру с Id
		Id int `json:"id"`
	}

	contentType := req.Header.Get("Content-Type") // Получаем тип заголовка входящего запроса
	mediatype, _, err := mime.ParseMediaType(contentType) // Анализируем значение типа мультимедиа и любые
	// необязательные параметры в соответствии с RFC 1521
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body) // С помощью функции NewDecoder из пакета json создается декодер для содержимого json файла.
	dec.DisallowUnknownFields() // DisallowUnknownFields вызывает ошибку, если входящий JSON содержит ключи, не соответствующие ни одному из них

	var rt RequestTask
	if err := dec.Decode(&rt); err != nil { // Если JSON недействительный выводим ошибку об этом
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := ts.store.CreateTask(rt.Text, rt.Tags, rt.Due)
	renderJSON(w, ResponseId{Id: id})
}

func (ts *taskServer) getAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all tasks at %s\n", req.URL.Path)

	allTasks := ts.store.GetAllTasks()
	renderJSON(w, allTasks)
}

func (ts *taskServer) deleteAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete all tasks at %s\n", req.URL.Path)
	ts.store.DeleteAllTasks()
}

func (ts *taskServer) getTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get task at %s\n", req.URL.Path)

	id, _ := strconv.Atoi(mux.Vars(req)["id"]) // Vars содержит переменные URI текущего запроса.
	task, err := ts.store.GetTask(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	renderJSON(w, task)
}

func (ts *taskServer) deleteTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete task at %s\n", req.URL.Path)

	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	err := ts.store.DeleteTask(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func (ts *taskServer) tagHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get tasks by tags at %s\n", req.URL.Path)

	tag := mux.Vars(req)["tag"]
	tasks := ts.store.GetTasksByTag(tag)
	renderJSON(w, tasks)
}

func (ts *taskServer) dueHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling tasks by due at %s\n", req.URL.Path)

	vars := mux.Vars(req)
	badRequestError := func() {
		http.Error(w, fmt.Sprintf("expect /due/<year>/<month>/<day>, got %v", req.URL.Path), http.StatusBadRequest)
		return
	}

	year, _ := strconv.Atoi(vars["year"])
	month, _ := strconv.Atoi(vars["month"])
	day, _ := strconv.Atoi(vars["day"])
	if month < int(time.January) || month > int(time.December) {
		badRequestError()
		return
	}

	tasks := ts.store.GetTasksByDueDate(year, time.Month(month), day)
	renderJSON(w, tasks)
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}