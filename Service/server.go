package main

import (
	myprotocol "MyPractice"
	"google.golang.org/grpc"
	//"net"
	//"context"
	//"MyPractice/MyDB"
	"fmt"
	_ "log"
	"net"
)

//var u myprotocol.ProdService

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	/*_, err := MyDB.DBInit("my.db")
	if err != nil {
		log.Fatal(err)
	}
	defer MyDB.DBClose()
	fmt.Println("数据库开打成功！")*/

	rpcServer := grpc.NewServer()
	myprotocol.RegisterProdServiceServer(rpcServer, &myprotocol.ProdService{})

	lis, _ := net.Listen("tcp", "10.30.61.210:9564")

	err := rpcServer.Serve(lis)
	if err != nil{
		fmt.Print(err)
	}
	fmt.Println("服务器启动！")




}

