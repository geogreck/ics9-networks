package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"lab1/src/proto"
	"net"
	"os"
)

// interact - функция, содержащая цикл взаимодействия с сервером.
func interact(conn *net.TCPConn) {
	defer conn.Close()
	encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		// Чтение команды из стандартного потока ввода
		fmt.Printf("command = ")
		var command string
		reader := bufio.NewReader(os.Stdin)
		buf, _, _ := reader.ReadLine()
		command = string(buf)

		// Отправка запроса.
		switch command {
		case "quit":
			sendRequest(encoder, "quit", nil)
			return
		case "insert html":
			var page proto.HtmlPage
			fmt.Printf("html page code = ")
			buf, _, _ := reader.ReadLine()
			page.Code = string(buf)
			sendRequest(encoder, "insert html", &page)
		case "parent node":
			sendRequest(encoder, "parent node", nil)
		case "first child":
			sendRequest(encoder, "first child", nil)
		case "last child":
			sendRequest(encoder, "last child", nil)
		case "prev sibling":
			sendRequest(encoder, "prev sibling", nil)
		case "next sibling":
			sendRequest(encoder, "next sibling", nil)
		case "current node":
			sendRequest(encoder, "current node", nil)
		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// Получение ответа.
		var resp proto.Response
		if err := decoder.Decode(&resp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		// Вывод ответа в стандартный поток вывода.
		switch resp.Status {
		case "ok":
			fmt.Printf("ok\n")
		case "failed":
			if resp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")
			} else {
				var errorMsg string
				if err := json.Unmarshal(*resp.Data, &errorMsg); err != nil {
					fmt.Printf("error: malformed data field in response\n")
				} else {
					fmt.Printf("failed: %s\n", errorMsg)
				}
			}
		case "result":
			if resp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")
			} else {
				var node proto.NodeForTransfer
				if err := json.Unmarshal(*resp.Data, &node); err != nil {
					fmt.Printf("error: malformed data field in response\n")
				} else {
					fmt.Printf("Name: %s\nAttributes: %v\nParent: %v\nFirst child: %v\nLast child: %v\n", node.Data, node.Attr, node.Parent, node.FirstChild, node.LastChild)
				}
			}
		default:
			fmt.Printf("error: server reports unknown status %q\n", resp.Status)
		}
	}
}

// sendRequest - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func sendRequest(encoder *json.Encoder, command string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(&proto.Request{Command: command, Data: &raw})
}

func main() {
	//buf := `<!DOCTYPE html><html lang="en"><head> <meta charset="UTF-8"> <meta http-equiv="X-UA-Compatible" content="IE=edge"> <meta name="viewport" content="width=device-width, initial-scale=1.0"> <title>Document</title></head><body> </body></html>`
	/* foo := "<div><p>1232</p>dsadsada</div>"
	res, _ := html.Parse(strings.NewReader(foo))
	fmt.Println(res.LastChild.LastChild.FirstChild.FirstChild.NextSibling) */
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, установка соединения с сервером и
	// запуск цикла взаимодействия с сервером.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else if conn, err := net.DialTCP("tcp", nil, addr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		interact(conn)
	}
}
