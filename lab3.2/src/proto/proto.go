package proto

// Config - параметры пира
type Config struct {
	Name       string           `json:"name"`
	Addr       string           `json:"addr"`
	KnownHosts []ConfigForShare `json:"knownHosts"`
}

// ConfigForShare - сокращенные параметры пира для хранения в клиенте других пиров
type ConfigForShare struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// Message - сообщение, которое пиры отправляют друг другу
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Sender  string `json:"sender"`
}
