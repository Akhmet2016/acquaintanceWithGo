package main

import (
	"net/http"
	"strconv"
	"time"
	"gin-rest-api/taskstore"
	"github.com/gin-gonic/gin"
)

type taskServer struct {
	store *taskstore.TaskStore
}

func NewTaskServer() *taskServer {
	store := taskstore.New()
	return &taskServer{store: store}
}

func main() {
	router := gin.Default() // Возвращает новый экземпляр Engine — основного типа данных Gin, который не только
	// играет роль маршрутизатора, но и даёт нам другой функционал. В частности, Default регистрирует базовое
	//ПО промежуточного уровня, используемое при восстановлении после сбоев и для логирования.
	server := NewTaskServer()

	router.POST("/task/", server.createTaskHandler)
	router.GET("/task/", server.getAllTasksHandler)
	router.DELETE("/task/", server.deleteAllTasksHandler)
	router.GET("/task/:id", server.getTaskHandler)
	router.DELETE("task/:id", server.deleteTaskHandler)
	router.GET("tag/:tag", server.tagHandler)
	router.GET("/due/:year/:month/:day", server.dueHandler)

	router.Run("localhost:8080")
}

func (ts *taskServer) createTaskHandler(c *gin.Context) { // В gin нет стандартных сигнатур HTTP-обработчиков Go,
	// вместо этого объект gin.Context, используется для анализа запроса и для формирования ответа
	type RequestTask struct {
		Text string `json:"text"`
		Tags []string `json:"tags"`
		Due time.Time `json:"due"`
	}

	var rt RequestTask
	if err := c.ShouldBindJSON(&rt); err != nil { // За разбор JSON-данных запроса отвечает ShouldBindJSON
		c.String(http.StatusBadRequest, err.Error())
	}

	id := ts.store.CreateTask(rt.Text, rt.Tags, rt.Due)
	c.JSON(http.StatusOK, gin.H{"Id": id}) // В Gin есть собственный механизм Context.JSON, который позволяет формировать JSON-ответы.
	// Теперь нам не нужно пользоваться «одноразовой» структурой для ID ответа. Вместо этого мы используем gin.H — псевдоним для map[string]interface{}
}

func (ts *taskServer) getAllTasksHandler(c *gin.Context) {
	allTasks := ts.store.GetAllTasks()
	c.JSON(http.StatusOK, allTasks)
}

func (ts *taskServer) deleteAllTasksHandler(c *gin.Context) {
	ts.store.DeleteAllTasks()
}

func (ts *taskServer) getTaskHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id")) // Gin позволяет обращаться к параметрам маршрута :id через Context.Params
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	task, err := ts.store.GetTask(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, task)
}

func (ts *taskServer) deleteTaskHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if err = ts.store.DeleteTask(id); err != nil {
		c.String(http.StatusNotFound, err.Error())
	}
}

func (ts *taskServer) tagHandler(c *gin.Context) {
	tag := c.Params.ByName("atg")
	tasks := ts.store.GetTasksByTag(tag)
	c.JSON(http.StatusOK, tasks)
}

func (ts *taskServer) dueHandler(c *gin.Context) {
	badRequestError := func() {
		c.String(http.StatusBadRequest, "expect /due/<year>/<month>/<day>, got %v", c.FullPath())
	}

	year, err := strconv.Atoi(c.Params.ByName("year"))
	if err != nil {
		badRequestError()
		return
	}

	month, err := strconv.Atoi(c.Params.ByName("month"))
	if err != nil || month < int(time.January) || month > int(time.December) {
		badRequestError()
		return
	}

	day, err := strconv.Atoi(c.Params.ByName("day"))
	if err != nil {
		badRequestError()
		return
	}

	tasks := ts.store.GetTasksByDueDate(year, time.Month(month), day)
	c.JSON(http.StatusOK, tasks)
}