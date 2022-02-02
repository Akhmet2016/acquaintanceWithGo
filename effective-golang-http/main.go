package main

import (
	"effective-golang-http/taskstore" // Подключаем наш пакет taskstore
	"encoding/json" // Пакет json реализует кодирование и декодирование JSON
	"fmt" // Пакет fmt реализует форматированный ввод-вывод
	"log" // Пакет для логирования
	"mime" // MIME — это аббревиатура английского многоцелевого почтового расширения Интернета:
	// MIME — это интернет-стандарт, который расширяет формат электронной почты для поддержки:
	// Текст в наборах символов, отличных от ASCII
	"net/http" // Пакет http предоставляет реализации HTTP-клиента и сервера
	"os" // Пакет os предоставляет независимый от платформы интерфейс к функциям операционной системы
	"strconv" // Пакет strconv реализует преобразования в строковые представления основных типов данных и обратно
	"strings" // Пакет реализуют простые функции для управления строками в кодировке UTF-8
	"time" // Пакет time предоставляет функциональные возможности для измерения и отображения времени
)

func main() {
	mux := http.NewServeMux() // Создает новый мультиплексор HTTP-запросов и присваивает его переменной mux.
	// Мультиплексор запросов сопоставляет URL-адрес входящих запросов со списком зарегистрированных путей
	// и вызывает соответствующий обработчик для пути.

	server := NewTaskServer() // Присваиваем переменной server функцию NewTaskServer - это конструктор для
	// нашего сервера, имеющего тип taskServer

	mux.HandleFunc("/task/", server.taskHandler) // регистрируем функцию-обработчик для пути /task/
	mux.HandleFunc("/tag/", server.tagHandler) // регистрируем функцию-обработчик для пути /tag/
	mux.HandleFunc("/due/", server.dueHandler) // регистрируем функцию-обработчик для пути /due/

	os.Setenv("SERVERPORT", "4000") // устанавливаем значение переменной среды, названной по ключу SERVERPORT.
	log.Fatal(http.ListenAndServe("localhost:"+os.Getenv("SERVERPORT"), mux)) // метод http.ListenAndServe() запускает сервер на порту 4000
}

type taskServer struct {
	store *taskstore.TaskStore
}

func NewTaskServer() *taskServer {
	store := taskstore.New() // Присваиваем структуру TaskStore из пакета taskstore, который содержит словарь
	// map[int]Task, данные при этом хранятся в памяти, айди записи с автоинкриментом, так же sync.Mutex, который
	// только блокирует и разблокирует.

	log.Printf("store %+v\n%#v\n", *store, *store, *store)
	return &taskServer{store: store} // Возварщаем структуру taskServer со store = структуре TaskStore
}

func (ts *taskServer) taskHandler(w http.ResponseWriter, req *http.Request) { // обработчик путей taskHandler
	if req.URL.Path == "/task/" { // Проверяем направлен ли запрос по пути /task/
		if req.Method == http.MethodPost { // Если метод запроса POST
			ts.createTaskHandler(w, req) // Вызываем обработчик пути createTaskHandler
		} else if req.Method == http.MethodGet { // Если метод запроса GET
			ts.getAllTasksHandler(w, req) // Вызываем обработчик пути getAllTasksHandler
		} else if req.Method == http.MethodDelete { // Если метод запроса DELETE
			ts.deleteAllTasksHandler(w, req) // Вызываем обработчик пути deleteAllTasksHandler
		} else { // Если наш сервер не реализует такой метод
			http.Error(w, fmt.Sprintf("Expect method GET, DELETE or POST at /task/, got %v", req.Method), http.StatusMethodNotAllowed)
			// Записываем в поток ответа ResponseWriter текст ошибки и ставим статус http StatusMethodNotAllowed (Метод не разрешен)
			return
		}
	} else {
		path := strings.Trim(req.URL.Path, "/") // Trim возвращает срез строки req.URL.Path с удаленными ведущими и конечными символами /
		pathParts := strings.Split(path, "/") // Split разбивает path на подстроки, разделенные символом /, и возвращает срез подстрок
		// между этими разделителями.

		if len(pathParts) < 2 { // Если длина среза меньше 2
			http.Error(w, "Expect /task/<id> in task handler", http.StatusBadRequest)
			// Записываем в поток ответа ResponseWriter текст ошибки и ставим статус http StatusBadRequest (Неверный запрос)
			return
		}
		id, err := strconv.Atoi(pathParts[1]) // Делаем числовое преобразования Atoi (string в int)
		if err != nil { // Обработка в случае ошибки
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Method == http.MethodDelete { // Если метод запроса DELETE
			ts.deleteTaskHandler(w, req, int(id)) // Вызываем обработчик пути deleteTaskHandler, передав туда айди
		} else if req.Method == http.MethodGet { // Если метод запроса GET
			ts.getTaskHandler(w, req, int(id)) // Вызываем обработчик пути getTaskHandler, передав туда айди
		} else {
			http.Error(w, fmt.Sprintf("Expect method GET or DELETE at /task/<id>, got %v", req.Method), http.StatusMethodNotAllowed)
			return
		}
	}
}

func (ts *taskServer) createTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling task create at %s\n", req.URL.Path) // Выводим лог о том, что произошёл запрос по пути...
	type RequestTask struct { // Объявляем новый тип, структуру RequestTask
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}

	type ResponseId struct { // Объявляем новый тип, структуру ResponseId
		Id int `json:"id"`
	}

	contentType := req.Header.Get("Content-Type") // Получаем тип контента
	mediatype, _, err := mime.ParseMediaType(contentType) // Анализируем значение типа мультимедиа и любые
	// необязательные параметры в соответствии с RFC 1521
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" { // Если тип не равен application/json, то выведится ошибка
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body) // С помощью функции NewDecoder из пакета json создается декодер для содержимого json файла.
	dec.DisallowUnknownFields() // DisallowUnknownFields вызывает ошибку, если входящий JSON содержит ключи, не соответствующие ни одному из них
	var rt RequestTask // Объявляем переменную с типом RequestTask
	if err := dec.Decode(&rt); err != nil { // Если JSON недействительный выводим ошибку об этом
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := ts.store.CreateTask(rt.Text, rt.Tags, rt.Due) // Вызываем CreateTask в пакете taskstore
	renderJSON(w, ResponseId{Id: id}) // вызываем renderJSON, передаём ResponseWriter и структуру с id
}

func (ts *taskServer) getAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all tasks at %s\n", req.URL.Path)

	allTasks := ts.store.GetAllTasks()
	renderJSON(w, allTasks)
}

func (ts *taskServer) getTaskHandler(w http.ResponseWriter, req *http.Request, id int) {
	log.Printf("handling get task at %s\n", req.URL.Path)

	task, err := ts.store.GetTask(id)
	if err != nil { // Если возвращается ошибка, то выводим её
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	renderJSON(w, task)
}

func (ts *taskServer) deleteTaskHandler(w http.ResponseWriter, req *http.Request, id int) {
	log.Printf("Handling delete task at %s\n", req.URL.Path)

	err := ts.store.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func (ts *taskServer) deleteAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Handling delete all tasks at %s\n", req.URL.Path)
	ts.store.DeleteAllTasks()
}

func (ts *taskServer) tagHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling tasks by tag at %s\n", req.URL.Path)

	if req.Method != http.MethodGet { // Если не метод GET
		http.Error(w, fmt.Sprintf("Expect method GET /tag/<tag>, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	path := strings.Trim(req.URL.Path, "/") // Убираем / по краям
	pathParts := strings.Split(path, "/") // Разделяем на срез по /
	if len(pathParts) < 2 { // Если длинна среза меньше 2
		http.Error(w, "Expect /tag/<tag> path", http.StatusBadRequest)
		return
	}
	tag := pathParts[1] // Значение tag является вторым элементом среза поэтому [1]

	tasks := ts.store.GetTaskByTag(tag)
	renderJSON(w, tasks)
}

func (ts *taskServer) dueHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Handling tasks by due at %s\n", req.URL.Path)

	if req.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Expect method GET /due/<date>, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	path := strings.Trim(req.URL.Path, "/")
	pathParts := strings.Split(path, "/")

	badRequestError := func() { // Присваиваем функцию обработки ошибки
		http.Error(w, fmt.Sprintf("Expect /due/<year>/<month>/<day>, got %v", req.URL.Path), http.StatusBadRequest)
	}

	if len(pathParts) != 4 { // Если запрос не состоит из 5 частей, то ошибка
		badRequestError()
		return
	}

	year, err := strconv.Atoi(pathParts[1]) // Присваиваем год, ковертируя строку в число
	if err != nil {
		badRequestError()
		return
	}
	month, err := strconv.Atoi(pathParts[2]) // Присваиваем месяц, ковертируя строку в число
	if err != nil || month < int(time.January) || month > int(time.December) { // Если вернулось с ошибкой
		// или месяц меньше чем 1, или больше чем 12, то ошибка
		badRequestError()
		return
	}
	day, err := strconv.Atoi(pathParts[3]) // Присваиваем день, ковертируя строку в число
	if err != nil {
		badRequestError()
		return
	}
	tasks := ts.store.GetTasksByDueDate(year, time.Month(month), day) // Заупскаем в пакете taskstore функцию GetTasksByDueDate,
	// передав туда год, месяц, день
	renderJSON(w, tasks)
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v) // Функция Marshal кодирует данные в формат JSON
	if err != nil { // Если возвращается ошибка, выводим её
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Устанавливаем в шапке тип
	w.Write(js) // Выводим этот json
}