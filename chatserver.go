package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

const BUFFERSIZE int = 1024

type User struct {
	Username string
	Login    bool
	Key      string
}

var buffer [BUFFERSIZE]byte

type response struct {
	Type    string
	To      string
	Message string
}

var allclient_conns = make(map[net.Conn]string)
var allloggedin_conns = make(map[net.Conn]interface{})
var lostclient = make(chan net.Conn)
var newclient = make(chan net.Conn)
var currentloggeduser User
var currentloggedusername string
var xuserlist []User
var userlist []string
var usernamelist string

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <port>\n", os.Args[0])
		os.Exit(0)
	}

	port := os.Args[1]
	if len(port) > 5 {
		fmt.Println("Invalid port value. Try again!")
		os.Exit(1)
	}

	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Cannot listen on port '" + port + "'!\n")
		os.Exit(2)
	}

	fmt.Println("ChatServer in GoLang developed by Vijayendra Kosigi, SECAD")
	fmt.Printf("ChatServer is listening on port '%s' ...\n", port)

	go func() {
		for {
			client_conn, _ := server.Accept()
			welcomemessage := fmt.Sprintf("New client is connected from '%s' Waiting for login!\n", client_conn.RemoteAddr().String())
			fmt.Println(welcomemessage)
			go userlogin(client_conn)
		}
	}()

	for {
		select {

		case client_conn := <-newclient:
			allclient_conns[client_conn] = client_conn.RemoteAddr().String()
			allloggedin_conns[client_conn] = currentloggedusername
			fmt.Println("Total number of active users at the moment is ", len(allloggedin_conns))
			if allloggedin_conns[client_conn] != "" {
				go client_goroutine(client_conn)
			}

		case client_conn := <-lostclient:
			go logout(client_conn)
		}

	}
}
func client_goroutine(client_conn net.Conn) {
	newuserfirstmsg := fmt.Sprintf("New user %s logged into Chat System  from %s. %s (from %d connections)", currentloggedusername, client_conn.RemoteAddr().String(), genuserlist(), len(userlist))
	fmt.Println(newuserfirstmsg)
	sendtoall([]byte(newuserfirstmsg))
	fmt.Printf("Total Connected Clients: %d\n", len(allloggedin_conns))
	go func() {
		for {
			byte_received, read_err := client_conn.Read(buffer[0:])
			if read_err != nil {
				lostclient <- client_conn
				return
			}
			fmt.Printf("\nReceived data: %s\n", buffer[0:byte_received])
			userrequestcalls(client_conn, buffer[0:byte_received])
		}
	}()
}

func userlogin(client_conn net.Conn) {
	byte_received, read_err := client_conn.Read(buffer[0:])
	if read_err != nil {
		fmt.Println("Error in receiving...")
		lostclient <- client_conn
		return
	}
	fmt.Printf("Got data : %s Expecting Login Data\n", buffer[0:byte_received])
	status, Username, message := checklogin(buffer[0:byte_received])
	if status {
		currentloggeduser = User{Username: Username, Login: true, Key: client_conn.RemoteAddr().String()}
		currentloggedusername = Username
		fmt.Println(currentloggeduser)
		newclient <- client_conn
		userlist = append(userlist, currentloggeduser.Username)
		xuserlist = append(xuserlist, currentloggeduser)
		usernamelist = usernamelist + ", " + currentloggeduser.Username
	} else {
		cantlogin := fmt.Sprintf("Authentication failed! Invalid username or password")
		client_conn.Write([]byte(cantlogin))
		go userlogin(client_conn)
	}
	fmt.Println(message)
}

func privatechat(client_conn net.Conn, receiver string, msg string) {
	for recvclient_conn, _ := range allloggedin_conns {
		if allloggedin_conns[recvclient_conn] == receiver {
			rx_user := recvclient_conn
			privmsg := fmt.Sprintf("%s: %s", allloggedin_conns[client_conn], msg)
			sendto(rx_user, []byte(privmsg))
		}
	}
}

func publicchat(client_conn net.Conn, publicchatmsg string) {
	pubmsg := fmt.Sprintf("\nPublic message from %s : %s", allloggedin_conns[client_conn], publicchatmsg)
	sendtoall([]byte(pubmsg))
	return
}
func userrequestcalls(client_conn net.Conn, data []byte) {
	var reply response
	err := json.Unmarshal(data, &reply)
	if err == nil {
		if reply.Type == "userlist" {
			client_conn.Write([]byte(genuserlist()))
			return
		} else if reply.Type == "public" {
			publicmsg := reply.Message
			publicchat(client_conn, publicmsg)
			return
		} else if reply.Type == "private" {
			privatechat(client_conn, reply.To, reply.Message)
			return
		} else if reply.Type == "exit" {
			lostclient <- client_conn
			return
		}

	}
}

func sendtoall(data []byte) {
	fmt.Printf("\nSent to all connected clients: %s\n", data)
	for u, _ := range allloggedin_conns {
		sendto(u, data)
	}
}

func genuserlist() string {
	var alluserlist []string
	for x, _ := range userlist {
		alluserlist = append(alluserlist, userlist[x])
	}
	fmt.Println("All the active users at the moment are: ")
	bytealluserlist := strings.Join(alluserlist, ",")
	return bytealluserlist
}

func sendto(client_conn net.Conn, data []byte) {
	_, write_err := client_conn.Write(data)
	if write_err != nil {
		fmt.Println("DEBUG>Error in sending to " + client_conn.RemoteAddr().String())
		return
	}
}

func checkaccount(Username string, Password string) bool {
	if Username == "kosigiv1" && Password == "hello1234" {
		return true
	} else if Username == "kosigiv2" && Password == "hello1234" {
		return true
	} else if Username == "testuser" && Password == "hello1234" {
		return true
	}
	return false

}

func checklogin(data []byte) (bool, string, string) {

	type Account struct {
		Username string
		Password string
	}

	var account Account
	err := json.Unmarshal(data, &account)
	if err != nil || account.Username == " " || account.Password == " " {
		fmt.Printf("JSON parsing error : %s\n", err)
		return false, " ", `[BAD LOGIN] Expected:  {"Username":"..","Password":".."}`
	}
	fmt.Printf("DEBUG>Got: account =%s\n", account)
	fmt.Printf("DEBUG>Got: username=%s,password=%s\n", account.Username, account.Password)
	if checkaccount(account.Username, account.Password) {
		return true, account.Username, "logged in successfully"
	}
	return false, "", "Invalid username or password"
}

func deleteuser(finaluserlist []string, name string, j int) []string {
	if len(finaluserlist) >= j {
		if finaluserlist[j] == name {
			finaluserlist = append(finaluserlist[:j], finaluserlist[j+1:]...)
			//break
		}
	}
	return finaluserlist
}

func nameandindex(client_conn net.Conn) (string, int) {
	var uname string
	var ind int
	for m, k := range xuserlist {
		if k.Key == client_conn.RemoteAddr().String() {
			fmt.Println("Username and index are found!!")
			uname = k.Username
			ind = m
		}
	}
	return uname, ind
}

func logout(client_conn net.Conn) []string {
	byemessage := fmt.Sprintf("\n%s client is disconnected", allloggedin_conns[client_conn])
	go sendtoall([]byte(byemessage))
	username, i := nameandindex(client_conn)
	userlist = deleteuser(userlist, username, i)
	fmt.Println("\nUpdated user list is: ")
	fmt.Println(userlist)
	go delete(allloggedin_conns, client_conn)
	client_conn.Close()
	return userlist
}
