package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-DB/proto"
	"os"
	//"sync"
	"google.golang.org/grpc/grpclog"
	"io"
)

const (
	address = "127.0.0.1:8077"
)

func StudyMethod(client pb.SelfManageClient) {
	fmt.Println("StudyMethod================简单模式================")
	//简单模式下直接调用方法就可以传递参数和获得返回值
	info := &pb.Study{Sid: 1, Age: 24, Name: "Deng", Telephone: "1888888", Address: "sz"}
	re, err := client.StudyAdd(context.Background(), info)
	if err != nil {
		fmt.Println("some error occur in StudyMethod", err)
	}
	fmt.Println("StudyMethod:", re.Sid, re.Information)
}

func GetHelloTestMethod(client pb.SelfManageClient) {
	fmt.Println("GetHelloTestMethod================客户端模式================")
	//此处相当于创建了一个GetHelloTest的通道。通过通道发送和接收数据。
	re, err := client.GetHelloTest(context.Background())
	if err != nil {
		fmt.Println("some error occur in gethellotest", err)
	}
	//初始化要发送给服务器端的数据，并发送（此处连续发送10次）
	for i := 0; i < 10; i++ {
		req := &pb.StudyRequest{Sid: 1, Information: "Biao"}
		re.Send(req) //客户端要先发送数据再接收
	}
	//使用for循环接收数据
	for {
		//先关闭send然后再接收数据。该方法内部调用了CloseSend和RecvMsg方法。但是后者协程不安全
		r, err2 := re.CloseAndRecv()
		if err2 == io.EOF {
			fmt.Println("recv done")
			return
		}
		if err2 != nil {
			fmt.Println("some error occur in gethellotest recv")
		}
		fmt.Println(r)
	}
}

func GetNxinMethod(client pb.SelfManageClient) {
	fmt.Println("GetNxinMethod================服务器端模式================")
	req := &pb.StudyRequest{Sid: 10, Information: "Biao"}
	//向服务器端以参数形式发送数据s，并创建了一个通道，通过通道来获取服务器端返回的流
	re, err := client.GetNxinMethod(context.Background(), req)
	if err != nil {
		fmt.Println("some error occur in GetNxinMethod")
	}
	for {
		//上面已经以参数的形式发送了数据，此处直接使用Recv接收数据。因是参数发送，所有不用使用CloseAndRecv
		v, e := re.Recv()
		if e == io.EOF {
			fmt.Println("recv done")
			return
		}
		if e != nil {
			fmt.Println("e not nil", e)
			return
		}
		fmt.Println("get nxin method recv", v)
	}
}

func GetStudyMethod(client pb.SelfManageClient) {
	fmt.Println("GetStudyMethod================双向模式================")
	re, err := client.GetStudy(context.Background())
	if err != nil {
		fmt.Println("get Study error", err)
	}
	inChan := make(chan *pb.StudyResponse, 1)
	outChan := make(chan struct{}, 1)
	inputChan := make(chan string, 1)
	go func() {
		fmt.Print("-> ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		req := &pb.StudyRequest{Sid: 2000, Information: text}
		re.Send(req)
		for {
			select {
			case in, ok := <-inChan:
				if !ok {
					return
				}
				fmt.Println(in.Sid, ":", in.Information)
				fmt.Print("-> ")
				reader := bufio.NewReader(os.Stdin)
				text, _ := reader.ReadString('\n')
				req := &pb.StudyRequest{Sid: 2000, Information: text}
				re.Send(req)
			case <-outChan:
				fmt.Println("client over")
				return
			}
		}
	}()

	for {
		in, err := re.Recv()
		if err == io.EOF {
			fmt.Println("client read done ")
			return
		}
		select {
		case inChan <- in:
		case <-outChan:
			fmt.Println("client over")
			return
		}
	}

	defer func() {
		close(inChan)
		close(outChan)
		close(inputChan)
	}()
}
func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatal("some error in dial")
	}

	client := pb.NewSelfManageClient(conn)
	waitChan := make(chan struct{})

	//StudyMethod(client)
	//GetHelloTestMethod(client)
	//GetNxinMethod(client)
	go GetStudyMethod(client)

	defer func() {
		conn.Close()
		close(waitChan)
	}()
	<-waitChan
}
