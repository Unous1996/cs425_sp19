package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"bufio" 
)

type Wrap struct {
	ip_address string
	sender_name string
	vector [] int
	message string
}

var (
	num_of_participants int
	localhost string
	local_ip_address string
	port_number string
	own_name string
)

var (
	read_map map[string]*net.TCPConn
	send_map map[string]*net.TCPConn
	remote_ip_2_name map[string]string
	conn_2_port_num map[*net.TCPConn]string
	has_sent_name map[*net.TCPConn]bool
	ip_2_vectorindex map[string]int
)

var (
	vm_addresses = []string{"172.22.156.52:4444","172.22.158.52:4444","172.22.94.61:4444","172.22.156.53:4444","172.22.158.53:4444",
		"172.22.94.62:4444","172.22.156.54:4444","172.22.158.54:4444","172.22.94.63:4444","172.22.156.55:4444"}
	vector = []int{0,0,0,0,0,0,0,0,0,0}
	start_chan chan bool
	holdback_queue = []Wrap{}
)

func printVector() {
	for i := 0; i < len(vector); i++ {
		fmt.Println("#vector[", i , "]=", vector[i])
	}
}

func printArbitiaryVector(vec []int, vecname string){
	for i := 0; i < len(vector); i++ {
		fmt.Println("#" + vecname + "[", i , "]=", vec[i])
	}
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
	s := strings.Split((parse),",")
	var result [] int
	for i := 0; i < len(vector); i++{
		temp, _ := strconv.Atoi(s[i])
		result = append(result, temp)
	}
	return result
}

func vector_accack(attacker []int, update int) {
	fmt.Println("#update = ", update)
	printVector()
	printArbitiaryVector(attacker, "attacker")
	for i := 0; i < len(vector); i++ {
		if(i == update) {
			vector[i] += 0
		} else {
			if(attacker[i] > vector[i]) {
				vector[i] = attacker[i]
			}
		}
	}
}

func deliver(received_vector []int, update int, deliver_string string){
	vector_accack(received_vector, update)
	fmt.Println(deliver_string)
}

func able_to_deliver(received_vector []int, source_index int) bool{
	result := true
	for i := 0; i < len(vector); i++{
		if(i == source_index){
			if(received_vector[i] != vector[i] + 1){
				return false;
			}
		} else {
			if(received_vector[i] != vector[i]){
				return false;
			}
		}
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

		recevied_string_spilt := strings.Split(string(buff[0:j]), ";")

		if(recevied_string_spilt[0] == "NAME"){
			fmt.Println("#" + recevied_string_spilt[1] + "(Address:" + conn.RemoteAddr().String() + ")"+ " has joined the chat")
			remote_ip_2_name[strings.Split(conn.RemoteAddr().String(),":")[0]] = recevied_string_spilt[1]
			continue
		}

		received_ip_address := recevied_string_spilt[1]
		received_name := recevied_string_spilt[2]
		received_message := recevied_string_spilt[3]
		received_vector := deserialize(recevied_string_spilt[0])
		deliver_string := received_name + ":" + received_message
		fmt.Println("#Before incrementing, your vector is")
		printVector()


		if(able_to_deliver(received_vector, ip_2_vectorindex[received_ip_address])){
			deliver(received_vector, ip_2_vectorindex[local_ip_address], deliver_string)
			fmt.Println("local_ip_address = ", local_ip_address)
			fmt.Println("ip_2_vectorindex[local_ip_address] = ", ip_2_vectorindex[local_ip_address])
			fmt.Println("#After incrementing, your vector is")
			printVector()

			for{
				again := false
				for it:= 0; it < len(holdback_queue); it++ {
					if(it > len(holdback_queue)-1){
						break
					}
					object := holdback_queue[it]
					if(able_to_deliver(object.vector, ip_2_vectorindex[object.ip_address])){
						deliver(object.vector, ip_2_vectorindex[object.ip_address], object.message)
						again = true;
					}
				}
				if(again == false){
					break;
				}
			}

		} else {
			fmt.Println("Failed to Deliver the Message")
			Temp := Wrap{received_ip_address, received_name, received_vector, received_message}
			holdback_queue = append(holdback_queue, Temp)
		}
	}
}

func multicast(name string)  {
	for{
		var msg string
		var send_string string
		
		in := bufio.NewReader(os.Stdin)
		msg, _ = in.ReadString('\n')

		fmt.Println("The message that you are about to deliver is:", msg)
		vector[ip_2_vectorindex[local_ip_address]] += 1
		fmt.Println("You are incrementing index:", ip_2_vectorindex[local_ip_address])
		fmt.Println("After Incrementing, your vector became:\n")
		printVector()

		send_vector := serialize(vector)
		send_string = send_vector + ";" + local_ip_address + ";" + name + ";" + msg
		b := []byte(send_string)

		for _, conn := range send_map {
			if conn.RemoteAddr().String() == localhost {
				continue
			}
			conn.Write(b)
		}

		/*
		for i := 0; i < len(vector); i++ {
			fmt.Println("vector = ", vector[i])
		}
		*/
	}
}

func multicast_name(name string){
	for{
		var msg string
		var send_string string
		
		msg = name
		send_string = "NAME" + ";" + msg
		b := []byte(send_string)

		for _, conn := range send_map {
			if conn.RemoteAddr().String() == localhost || has_sent_name[conn]{
				continue
			}
			conn.Write(b)
			has_sent_name[conn] = true
		}

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
	for i := int64(0); i < 10; i++{
		if vm_addresses[i] != localhost {
			fmt.Println("Registering address:", vm_addresses[i])
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
	own_name = os.Args[1]
	port_number = os.Args[2]
	
	read_map = make(map[string]*net.TCPConn)
	send_map = make(map[string]*net.TCPConn)
	remote_ip_2_name = make(map[string]string)
	conn_2_port_num = make(map[*net.TCPConn]string)
	ip_2_vectorindex = make(map[string]int)
	has_sent_name = make(map[*net.TCPConn]bool)

	ip_2_vectorindex = map[string]int{
		"172.22.156.52": 0,
		"172.22.158.52": 1,
		"172.22.94.61": 2,
		"172.22.156.53": 3,
		"172.22.158.53": 4,
		"172.22.94.62": 5,
		"172.22.156.54": 6,
		"172.22.158.54": 7,
		"172.22.94.63": 8,
		"172.22.156.55": 9,
	}

	addrs, err := net.InterfaceAddrs()
    	if err != nil {
        	fmt.Println(err)
        	os.Exit(1)
    	}
    
	for _, address := range addrs {
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
			local_ip_address = ipnet.IP.String()
                fmt.Println("The local ip address is:", ipnet.IP.String())
            }
        }
    }
	
	//Listen on a port that we specified

	localhost = local_ip_address + ":" +port_number
	fmt.Println("Your assigned index is:", ip_2_vectorindex[local_ip_address])
	fmt.Println("#Start server...")
	go start_server(port_number)

	time.Sleep(5 * time.Second)

	fmt.Println("#Start client...")
	go start_client(num_of_participants)

	go multicast_name(own_name)
	go multicast(own_name)
	<-start_chan
}
