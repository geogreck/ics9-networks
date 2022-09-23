package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"lab1/src/proto"
	"net"
	"strings"

	log "github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
)

// Делает копию данного узла, содержащую только важную информацию для передачи клиенту
func createNodeForTransfer(node *html.Node) proto.NodeForTransfer {
	buf := proto.NodeForTransfer{
		Parent:    node.Parent.Data,
		Data:      node.Data,
		Namespace: node.Namespace,
		Attr:      node.Attr,
	}
	if node.FirstChild != nil {
		buf.FirstChild = node.FirstChild.Data
	} else {
		buf.FirstChild = "none"
	}
	if node.LastChild != nil {
		buf.LastChild = node.LastChild.Data
	} else {
		buf.LastChild = "none"
	}
	return buf
}

// Client - состояние клиента.
type Client struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	page   *html.Node
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case "quit":
		client.respond("ok", nil)
		return true
	case "insert html":
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else {
			var page proto.HtmlPage
			if err := json.Unmarshal(*req.Data, &page); err != nil {
				errorMsg = "malformed data field"
			} else {
				client.page, err = html.Parse(strings.NewReader(page.Code))
				if err != nil {
					errorMsg = "Parsing error, incorrect html page"
				} else {
					client.logger.Info("uploaded html page")
					client.respond("ok", nil)
					client.page = client.page.FirstChild
				}
			}
		}
		if errorMsg != "" {
			client.logger.Error("parsing failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "parent node":
		errorMsg := ""
		node := client.page.Parent
		if node == nil {
			errorMsg = "missing parent node"
		} else {
			client.page = node
			client.logger.Info("Found and returned parent node")
		}
		if errorMsg == "" {
			buf := createNodeForTransfer(node)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "first child":
		errorMsg := ""
		node := client.page.FirstChild
		if node == nil {
			errorMsg = "missing child node"
		} else {
			client.page = node
			client.logger.Info("Found and returned first child node")
		}
		if errorMsg == "" {
			buf := createNodeForTransfer(node)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "current node":
		errorMsg := ""
		fmt.Println(client.page)
		if client.page == nil {
			errorMsg = "missing node"
		}
		if errorMsg == "" {
			client.logger.Info("returned cur node")
			buf := createNodeForTransfer(client.page)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "last child":
		errorMsg := ""
		node := client.page.LastChild
		if node == nil {
			errorMsg = "missing child node"
		} else {
			client.page = node
			client.logger.Info("Found and returned last child node")
			fmt.Println(*node.LastChild)
		}
		if errorMsg == "" {
			buf := createNodeForTransfer(node)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "prev sibling":
		errorMsg := ""
		node := client.page.PrevSibling
		if node == nil {
			errorMsg = "missing prev sibling"
		} else {
			client.page = node
			client.logger.Info("Found and returned prev sibling node")
		}
		if errorMsg == "" {
			buf := createNodeForTransfer(node)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "next sibling":
		errorMsg := ""
		node := client.page.NextSibling
		if node == nil {
			errorMsg = "missing next sibling"
		} else {
			client.page = node
			client.logger.Info("Found and returned next sibling node")
		}
		if errorMsg == "" {
			buf := createNodeForTransfer(node)
			client.respond("result", &buf)
		} else {
			client.logger.Error("calculation failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{Status: status, Data: &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).serve()
				}
			}
		}
	}
}
