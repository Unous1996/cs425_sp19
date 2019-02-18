package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var num_of_participants int
var ConnMap map[string]*net.TCPConn
var current_number_of_ports_registered int64
var collected_ports string
var ch_read_err chan int = make(chan int)
var port_register = ":7998"

func checkErr(err error) int {
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("用户退出")
			return 0
		}
		fmt.Println("发生错误")
		return -1
	}
	return 1
}

func readPortString(finished chan int, conn *net.TCPConn){
	buff := make([]byte, 256)

	j, err := conn.Read(buff)
	if err != nil {
		ch_read_err <- 1
	} else {
		fmt.Println("Other Ports are:", string(buff[0:j]))
	}

	finished <- 1
}

func registerPort(tcpConn *net.TCPConn, capacity int64){
	data := make([]byte, 256) 
	total, err := tcpConn.Read(data) 
	if err != nil {
		fmt.Println(string(data[:total]), err)
	} else {
		temp_port := string(data[:5])
		fmt.Println("Received Port Number:", temp_port)
		current_number_of_ports_registered += 1	
		collected_ports = collected_ports + temp_port	
	}
    
	flag := checkErr(err)
	if flag == 0 {
		fmt.Println("接收时发生错误")
	}

	if(current_number_of_ports_registered == capacity - 1){
		print("First time reached here")
		for _, conn := range ConnMap {
			conn.Write([]byte(collected_ports))
		}
		fmt.Println("Reached Here:", current_number_of_ports_registered)
	}
}

func main(){
	if len(os.Args) != 4 {
		fmt.Println(os.Stderr, "Incorrect number of parameters")
		os.Exit(1)
	}

	finished := make(chan int)
	num_of_participants,_ := strconv.ParseInt(os.Args[3],0,64)
	port_number := os.Args[2]
	port_number = ":" + port_number

	if(port_number == ":7998"){
		tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:7998")
		tcpListen, _ := net.ListenTCP("tcp", tcpAddr)
		ConnMap = make(map[string]*net.TCPConn)
		//Listen for TCP networks
		var i int64;
		go func(){
			for i= 0; i < num_of_participants-1; i++ {
				tcpConn, _ := tcpListen.AcceptTCP()
				defer tcpConn.Close()
				ConnMap[tcpConn.RemoteAddr().String()] = tcpConn
				fmt.Println("连接客户端信息：", tcpConn.RemoteAddr().String())
				registerPort(tcpConn, num_of_participants)
			}
		}()
		time.Sleep(5*time.Second)
	} else {
		TcpAdd, _ := net.ResolveTCPAddr("tcp", port_register)
		conn, err := net.DialTCP("tcp", nil, TcpAdd)

		if err != nil {
			fmt.Println("The servers is currently closed")
			os.Exit(1)
		}

		sendbyte := []byte(port_number)
		conn.Write(sendbyte)

		defer conn.Close()
		fmt.Println("Create A goroutine for reading port string")
		go readPortString(finished, conn)
		<-finished
	}

	fmt.Println("READY")
}