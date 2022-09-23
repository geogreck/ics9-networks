package proto

import "encoding/json"

// Request -- запрос клиента к серверу.
type Request struct {
	// Поле Command может принимать три значения:
	// * "quit" - прощание с сервером (после этого сервер рвёт соединение);
	// * "calculate" - передача новой задачи на сервер;
	Command string `json:"command"`

	Data *json.RawMessage `json:"data"`
}

// Response -- ответ сервера клиенту.
type Response struct {
	// Поле Status может принимать три значения:
	// * "ok" - успешное выполнение команды "quit";
	// * "failed" - в процессе выполнения команды произошла ошибка;
	// * "result" - максимальная высота вычислена.
	Status string `json:"status"`

	// Если Status == "failed", то в поле Data находится сообщение об ошибке.
	// Если Status == "result", в поле Data должна лежать высота
	// В противном случае, поле Data пустое.
	Data *json.RawMessage `json:"data"`
}

// Task -- условие задачи для вычисления сервером
type Task struct {
	// Угол от горизонта в градусах
	Angle float64 `json:"angle"`

	// Общая скорость тела в начале движения
	Velocity float64 `json:"Velocity"`
}
