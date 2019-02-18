package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	num_of_participants int
	conn_map map[*net.TCPConn]string //for receving, not for sending
	send_map map[string]*net.TCPConn
	vm_addresses = []string{"10.192.137.227:7100","10.192.137.227:7200","10.192.137.227:7300","10.192.137.227:7400","10.192.137.227:7500",
		"10.192.137.227:7600","10.192.137.227:7700","10.192.137.227:7800","10.192.137.227:7900","10.192.137.227:8000"}
	name string
	start_chan chan bool
)


func checkErr(err error) int {
	if err != nil {
		if err.Error() == "EOF" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return 1
}

func readMessage(conn *net.TCPConn){

	buff := make([]byte, 256)
	for {
		j, err := conn.Read(buff)

		if j > 0 {
			s := strings.Split(string(buff[0:j]), ":")
			user := s[0]

			//Check user leave or error happen
			flag := checkErr(err)
			if flag == 0 {
				fmt.Println(user + " has left")
				break
			}
			fmt.Printf("%s\n", buff[0:j])
		}
	}
}

func sendMessage(name string)  {


	for{
		var msg string
		fmt.Println("Please Input the message to be send:")
		fmt.Scanln(&msg)
		b := []byte(name + ": " + msg)
		
		for addr, conn :=  range send_map{
			fmt.Println("About to send message")
			_, err := conn.Write(b)
			flag := checkErr(err)
 			
 			if err != nil{
 				fmt.Println("Error occured while sending message")
 			}

			if flag == 0 {
				fmt.Println("Flag = 0")
				fmt.Println("occured when delivering to the following address", addr)
			}
		}

	}
}

func start_server(num_of_participants int64, port_num string){

	//Listen on a port that we specified
	localhost := "10.192.137.227:" + port_num
	tcp_addr, _ := net.ResolveTCPAddr("tcp", localhost)
	tcp_listen, err := net.ListenTCP("tcp", tcp_addr)

	if err != nil {
		fmt.Println("Failed to listen on " + port_num)
	}

	fmt.Println("Start listening on " + port_num)
	// Accept Tcp connection from other VMs
	for {

		conn, _ := tcp_listen.AcceptTCP()
		defer conn.Close()
		conn_map[conn] = conn.RemoteAddr().String()
		go readMessage(conn)
	}

}

func start_client(num_of_participants int64, port_num string){

	localhost := "10.192.137.227:" + port_num

	fmt.Print("Please enter your name：")
	fmt.Scanln(&name)
	fmt.Println("Your name is：", name)

	//Create TCP connection to other VMs
	for i := int64(0); i < num_of_participants; i++{
		if vm_addresses[i] != localhost {
			tcp_add, _ := net.ResolveTCPAddr("tcp", vm_addresses[i])
			conn, err := net.DialTCP("tcp", nil, tcp_add)
			if err != nil {
				fmt.Println("Service unavailable on " + vm_addresses[i])
				continue
			}
			send_map[vm_addresses[i]] = conn
			defer conn.Close()
			// go function for chat, finished later
		}
	}

	go sendMessage(name)

	fmt.Println("Ready")
	<-start_chan
}

func main(){
	if len(os.Args) != 4 {
		fmt.Println(os.Stderr, "Incorrect number of parameters")
		os.Exit(1)
	}

	num_of_participants,_ := strconv.ParseInt(os.Args[3],0,64)
	port_number := os.Args[2]
	conn_map = make(map[*net.TCPConn]string)
	send_map = make(map[string]*net.TCPConn)
	go start_server(num_of_participants, port_number)

	time.Sleep(5 * time.Second)

	go start_client(num_of_participants, port_number)
	<-start_chan
}

