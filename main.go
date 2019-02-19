package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"strings"
)

var (
	num_of_participants int
	localhost string
	local_ip_address string
	port_number string

	read_map map[string]*net.TCPConn
	send_map map[string]*net.TCPConn
	remote_ip_2_name map[string]string
	conn_2_port_num map[*net.TCPConn]string
	port_2_vector_index map[string]int
	vm_addresses = []string{"10.192.137.227:7100","10.192.137.227:7200","10.192.137.227:7300","10.192.137.227:7400","10.192.137.227:7500",
		"10.192.137.227:7600","10.192.137.227:7700","10.192.137.227:7800","10.192.137.227:7900","10.192.137.227:8000"}
	vector = []int{0,0,0,0,0,0,0,0,0,0}
	start_chan chan bool

)

type standard_message struct {
	ip_address string
	sender_name string
	vector [] int
	message string
}

func checkErr(err error) int {
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println(err)
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return 1
}

func serialize(vec []int) string{
	result := "["
	for i := 0 ; i < len(vec) ; i++{
		result += ","
		result += strconv.Itoa(vec[i])
	}
	result += ",]"
	return result
}

func deserialize(str string) []int{
	parse := str[2:len(str)-2]
	fmt.Println("parse = ", parse)
	s := strings.Split((parse),",")
	var result [] int
	for i := 0; i < len(vector); i++{
		temp, _ := strconv.Atoi(s[i])
		result = append(result, temp)
	}
	return result
}

func readMessage(conn *net.TCPConn){

	buff := make([]byte, 256)
	for {
		j, err := conn.Read(buff)

		//Check user leave or error happen
		flag := checkErr(err)
		if flag == 0 {
			s := strings.Split(conn.RemoteAddr().String(),":")
			remote_ip := s[0]
			fmt.Println(remote_ip_2_name[remote_ip] + " has left")
			break
		}

		received_string := (string(buff[0:j]))
		recevied_string_spilt := strings.Split(string(buff[0:j]), ":")
		received_vector := deserialize(recevied_string_spilt[0])

		for i := 0; i < len(received_vector); i++{
			fmt.Println("received_vector = ", received_vector[i])
		}

		fmt.Println(received_string)
	}
}

func multicast(name string)  {
	for{
		var msg string
		var send_string string
		fmt.Scanln(&msg)

		send_vector := serialize(vector)
		fmt.Println("send_vector = ", send_vector)
		send_string = send_vector + ";" + local_ip_address + ";" + name + ";" + msg
		b := []byte(send_string)

		for _, conn := range send_map {
			if conn.RemoteAddr().String() == localhost {
				continue
			}
			conn.Write(b)
		}

		vector[port_2_vector_index[port_number]] += 1
		//fmt.Println(vector[port_2_vector_index[port_number]])
	}
}

func start_server(port_num string){

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
		conn_2_port_num[conn] = port_num
		read_map[conn.RemoteAddr().String()] = conn
		go readMessage(conn)
	}
}

func start_client(num_of_participants int64){

	//Create TCP connection to other VMs
	for i := int64(0); i < num_of_participants; i++{
		if vm_addresses[i] != localhost {
			tcp_add, _ := net.ResolveTCPAddr("tcp", vm_addresses[i])
			conn, err := net.DialTCP("tcp", nil, tcp_add)
			if err != nil {
				fmt.Println("Service unavailable on " + vm_addresses[i])
				continue
			}
			defer conn.Close()
			send_map[conn.RemoteAddr().String()] = conn
		}
	}

	fmt.Println("Ready")
	<-start_chan
}

func main(){
	if len(os.Args) != 4 {
		fmt.Println(os.Stderr, "Incorrect number of parameters")
		os.Exit(1)
	}

	num_of_participants,_ := strconv.ParseInt(os.Args[3],0,64)
	name := os.Args[1]
	port_number = os.Args[2]
	
	read_map = make(map[string]*net.TCPConn)
	send_map = make(map[string]*net.TCPConn)
	remote_ip_2_name = make(map[string]string)
	conn_2_port_num = make(map[*net.TCPConn]string)
	port_2_vector_index = make(map[string]int)
	port_2_vector_index = map[string]int{
		"7100": 0,
		"7200": 1,
		"7300": 2,
		"7400": 3,
		"7500": 4,
		"7600": 5,
		"7700": 6,
		"7800": 7,
		"7900": 8,
		"8000": 9,
	}

	remote_ip_2_name = map[string]string{
		"10.192.137.227": "Alice",
	}

	//Listen on a port that we specified
	local_ip_address = "10.192.137.227:"
	localhost = local_ip_address + port_number

	fmt.Println("Start server...")
	go start_server(port_number)

	time.Sleep(5 * time.Second)

	fmt.Println("Start client...")
	go start_client(num_of_participants)

	go multicast(name)
	<-start_chan
}