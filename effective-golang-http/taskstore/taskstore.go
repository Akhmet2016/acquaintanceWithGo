package taskstore

import (
	"fmt"
	"sync" // Пакет предоставляет базовые примитивы синхронизации, такие как блокировки взаимного исключения.
	"time"
)

type Task struct { // Объявляем структуру Task
	Id int `json:"id"`
	Text string `json:"text"`
	Tags []string `json:"tags"`
	Due time.Time `json:"due"`
}

type TaskStore struct { // Объявляем структуру TaskStore
	sync.Mutex
	tasks map[int]Task
	nextId int
}

func New() *TaskStore {
	ts := &TaskStore{} // Присваиваем структуру TaskStore
	ts.tasks = make(map[int]Task) // Присваиваем пустую мапу tasks, которая будет хранить всю информацию по таскам
	ts.nextId = 0 // Задаём начальный индекс равный 0
	return ts // Возвращаем структуру
}

func (ts *TaskStore) CreateTask(text string, tags []string, due time.Time) int {
	ts.Lock() // Вызван Lock, поэтому только одна go-процедура за раз
	// может получить доступ к TaskStore
	defer ts.Unlock() // defer нужен, чтобы удостовериться, что мьютекс будет освобожден

	task := Task{ // Инициализируем переменную task с типом Task, присваивая значения
		Id:   ts.nextId,
		Text: text,
		Due:  due,
	}

	task.Tags = make([]string, len(tags)) // Присваиваем ключу Tags срез типа string, длинной, равной длинне среза tags
	copy(task.Tags, tags) // Копируем значения в целевой срез из исходного

	ts.tasks[ts.nextId] = task // Записываем с store task с айди
	ts.nextId++ // Увиличиваем айди в store на единицу
	return task.Id
}

func (ts *TaskStore) GetTask(id int) (Task, error) {
	ts.Lock()
	defer ts.Unlock()

	t, ok := ts.tasks[id] // Забираем task по айди
	if ok {
		return t, nil
	} else { // Если не нашли, возвращаем пустой Task и ошибку с причиной
		return Task{}, fmt.Errorf("Task with id = %d not found", id)
	}
}

func (ts *TaskStore) DeleteTask(id int) error {
	ts.Lock()
	defer ts.Unlock()

	if _, ok := ts.tasks[id]; !ok { // Если нет task с айди, котрый заслан для удаления, выводим ошибку
		return fmt.Errorf("Task with id = %d not found", id)
	}

	delete(ts.tasks, id) // Удаляет элемент с ключом id из карты ts.tasks
	return nil
}

func (ts *TaskStore) DeleteAllTasks() error {
	ts.Lock()
	defer ts.Unlock()

	ts.tasks = make(map[int]Task) // В случае полного удаления, просто присваиваем ts.tasks пустую карту с типом Task
	return nil
}

func (ts *TaskStore) GetAllTasks() []Task {
	ts.Lock()
	defer ts.Unlock()

	allTasks := make([]Task, 0, len(ts.tasks)) // Инициализируем срез allTasks типа Task длинной 0 и вместимостью равной
	// длинне мапы ts.tasks
	for _, task := range ts.tasks { // Бежим по ts.tasks
		allTasks = append(allTasks, task) // Добавления новых элементы к срезу allTasks с помощью append
	}
	return allTasks
}

func (ts *TaskStore) GetTaskByTag(tag string) []Task {
	ts.Lock()
	defer ts.Unlock()

	var tasks []Task // Объявляем пустой срез типа Task

	taskloop: // Метка, чтобы можно было переметить на следующую цикл цикла
		for _, task := range ts.tasks { // Бежим по таскам
			for _, taskTag := range task.Tags { // Бежим по тегам таска
				if taskTag == tag { // Если тег равен тому, что пришел
					tasks = append(tasks, task) // Записывем тег
					continue taskloop // Переходим на верхний for, потому что этот таска уже подходит
				}
			}
		}
		return tasks
}

func (ts *TaskStore) GetTasksByDueDate(year int, month time.Month, day int) []Task {
	ts.Lock()
	defer ts.Unlock()

	var tasks []Task

	for _, task := range ts.tasks { // Бежим по таскам
		y, m, d := task.Due.Date() // Получаем день, месяц и год записи в сторе
		if y == year && m == month && d == day { // Если все элементы даты совпадают, то выводим,
			// учитываем, что месяц выводится в формате имени
			tasks = append(tasks, task)
		}
	}

	return tasks
}