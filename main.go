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
	ip_2_vectorindex map[string]int
	name_2_vector_index map[string]int
)

var (
	vm_addresses = []string{"10.180.129.254:7100","10.180.129.254:7200","10.180.129.254:7300","10.180.129.254:7400","10.180.129.254:7500",
		"10.180.129.254:7600","10.180.129.254:7700","10.180.129.254:7800","10.180.129.254:7900","10.180.129.254:8000"}
	vector = []int{0,0,0,0,0,0,0,0,0,0}
	start_chan chan bool
	holdback_queue = []Wrap{}
)

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
		received_ip_address := recevied_string_spilt[1]
		received_name := recevied_string_spilt[2]
		received_message := recevied_string_spilt[3]
		received_vector := deserialize(recevied_string_spilt[0])
		deliver_string := received_name + ":" + received_message

		if(able_to_deliver(received_vector, name_2_vector_index[received_name])){
			deliver(received_vector, name_2_vector_index[own_name], deliver_string)

			for{
				again := false
				for it:= 0; it < len(holdback_queue); it++ {
					if(it > len(holdback_queue)-1){
						break
					}
					object := holdback_queue[it]
					if(able_to_deliver(object.vector, name_2_vector_index[object.sender_name])){
						deliver(object.vector, name_2_vector_index[object.sender_name], object.message)
						again = true;
					}
				}
				if(again == false){
					break;
				}
			}

		} else {
			Temp := Wrap{received_ip_address, received_name, received_vector, received_message}
			holdback_queue = append(holdback_queue, Temp)
		}
		for i := 0; i < len(vector); i++ {
			fmt.Println("vector = ", vector[i])
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
		vector[ip_2_vectorindex[port_number]] += 1
		send_vector := serialize(vector)
		send_string = send_vector + ";" + local_ip_address + ";" + name + ";" + msg
		b := []byte(send_string)

		for _, conn := range send_map {
			if conn.RemoteAddr().String() == localhost {
				continue
			}
			conn.Write(b)
		}


		for i := 0; i < len(vector); i++ {
			fmt.Println("vector = ", vector[i])
		}
		//fmt.Println(vector[ip_2_vectorindex[port_number]])
	}
}

func start_server(port_num string){

	tcp_addr, _ := net.ResolveTCPAddr("tcp", localhost)
	tcp_listen, err := net.ListenTCP("tcp", tcp_addr)

	if err != nil {
		fmt.Println("#Failed to listen on " + port_num)
	}

	fmt.Println("#Start listening on " + port_num)
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
				fmt.Println("#Service unavailable on " + vm_addresses[i])
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
	name_2_vector_index = make(map[string]int)

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

	name_2_vector_index = map[string]int{
		"Alice": 0,
		"Bob": 1,
		"Cindy": 2,
		"Dick": 3,
		"Edgar": 4,
		"Fred": 5,
		"George": 6,
		"Henry": 7,
		"Ian": 8,
		"Jim": 9,
	}

	remote_ip_2_name = map[string]string{
		"10.180.129.254": "Alice",
	}

	//Listen on a port that we specified
	local_ip_address = "10.180.129.254:"
	localhost = local_ip_address + port_number

	fmt.Println("Start server...")
	go start_server(port_number)

	time.Sleep(5 * time.Second)

	fmt.Println("Start client...")
	go start_client(num_of_participants)

	go multicast(own_name)
	<-start_chan
}