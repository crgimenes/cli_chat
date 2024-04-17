package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Fabriciope/cli_chat/client/controller"
	"github.com/Fabriciope/cli_chat/client/cui"
	"github.com/Fabriciope/cli_chat/pkg/escapecode"
	"github.com/Fabriciope/cli_chat/pkg/shared/dto"
)

// TODO: colocar como variaveis globais no container
const (
	remoteIp   = "localhost"
	remotePort = 5000

	localIp   = "localhost"
	localPort = 3000
)

type User struct {
	connection   *net.TCPConn
	inputScanner *bufio.Scanner
	controller   *controller.Controller
	cui          cui.CUIInterface
	loggedIn     *bool // TODO: pensar em usar mutex ao alterar e acessar user.loggedIn
}

func NewUser(cui cui.CUIInterface) (*User, error) {
	remoteAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", remoteIp, remotePort))
	//localAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", localIp, localPort))
	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	loggedIn := false
	return &User{
		connection:   conn,
		inputScanner: bufio.NewScanner(os.Stdin),
		controller:   controller.NewController(conn, cui, &loggedIn),
		cui:          cui,
		loggedIn:     &loggedIn,
	}, nil
}

func (user *User) CloseConnection() error {
	return user.connection.Close()
}

func (user *User) InitChat() {
	defer user.CloseConnection()

	go user.cui.InitConsoleUserInterface()

	// TODO: colocar a logica de fazer o login para o main.go
	user.login()

	go user.listenToServer()
	user.listenToInput()
}

func (user *User) login() {
	for user.inputScanner.Scan() {
		username := strings.Trim(user.inputScanner.Text(), " ")
		if username == "" {
			user.cui.PrintLine(
				cui.MakeLine(&cui.Line{
					Info:      "login error:",
					InfoColor: escapecode.Red,
					Text:      "empty username",
				}))

			continue
		}

		loginWithCommand := ":login " + username
		user.controller.HandleInput(loginWithCommand)

		response, err := user.awaitResponseFromServer()
		if err != nil {
			user.cui.PrintLineForInternalError(err.Error())
			continue
		}

		user.controller.HandleResponse(response)
		if !*user.loggedIn {
			continue
		}

		return
	}
}

func (user *User) listenToServer() {
	for {
		var buf = make([]byte, 1024)
		n, err := user.connection.Read(buf)
		if err != nil {
			user.cui.PrintLineAndExit(1, cui.Line{
				Info:      "error from server:",
				InfoColor: escapecode.Red,
				Text:      "connection to the server was lost",
				TextColor: escapecode.Yellow,
			})
		}

		var responseFromServer dto.Response
		if err = json.Unmarshal(buf[:n], &responseFromServer); err != nil {
			return
		}

		user.controller.HandleResponse(responseFromServer)
	}
}

func (user *User) listenToInput() {
	for user.inputScanner.Scan() {
		if user.inputScanner.Err() != nil {
			return
		}

		input := strings.Trim(user.inputScanner.Text(), " ")
		if input == "" {
			user.cui.RedrawTypingBox()
			continue
		}

		user.controller.HandleInput(input)
		user.cui.RedrawTypingBox()
	}
}

func (user *User) awaitResponseFromServer() (responseFromServer dto.Response, err error) {
	var buf = make([]byte, 1024)
	n, err := user.connection.Read(buf)
	if err != nil {
		return
	}

	if err = json.Unmarshal(buf[:n], &responseFromServer); err != nil {
		return
	}

	return
}
