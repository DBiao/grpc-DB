package main

import (
	"bufio"
	"fmt"
	_ "fmt"
	"golang.org/x/net/context"
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	pb "grpc-DB/proto"
	"io"
	"log"
	"net"
	"os"
)

const (
	PORT = ":8077"
)

type SServer struct{}

//简单模式
func (ss *SServer) StudyAdd(ctx context.Context, in *pb.Study) (*pb.StudyResponse, error) {
	fmt.Println("StudyAdd================简单模式================")
	//简单模式不需要调用recv和send。数据直接通过方法的参数和返回值来接收发送。
	fmt.Println(in.Sid)
	resp := &pb.StudyResponse{Sid: 1000, Information: "Deng"}
	return resp, nil
}

//客户端流模式
func (ss *SServer) GetHelloTest(stream pb.SelfManage_GetHelloTestServer) error {
	fmt.Println("GetHelloTest================客户端模式================")
	//服务器端的代码stream是SelfManage_GetHelloTestServer类型，这种类型只有recv和sendandclose两种方法接收和发送并关闭
	//而客户端跟着相反是send和closeandrecv两种方法。
	//客户端流模式要先接收然后再发送。
	//使用for循环接收数据，因为接收的是流模式的StudyRequest类型，所以每次循环得到的是一个StudyRequest类型
	for {
		in, err := stream.Recv() //通过for循环接收客户端传来的流，该方法协程不安全。
		//必须把判断EOF的条件放在nil前面，因为如果获取完数据err会返回EOF类型
		if err == io.EOF { //在读取完所有数据之后会返回EOF来执行该条件。
			fmt.Println("read done")
			//发送（只在读取完数据后发送了一次）
			resp := &pb.StudyResponse{Sid: 1000, Information: "Deng"}
			//发送并关闭。该方法的返回值是error类型。但客户端并未得到返回值，只收到sendandclose发送的数据。
			//注意：该方法内部调用的SendMsg方法，但是该方法协程不安全。不同协程可以同时调用会出现问题。
			return stream.SendAndClose(resp)
		}
		if err != nil {
			fmt.Println("ERR IN GetHelloTest", err)
			return err
		}
		//打印出每次接收的数据
		fmt.Println("Recieved information:", in)
	}
}

//服务器端流模式
func (ss *SServer) GetNxinMethod(in *pb.StudyRequest, stream pb.SelfManage_GetNxinMethodServer) error {
	fmt.Println("GetNxinMethod================服务器端模式================")
	fmt.Println(in.Sid)
	resp := &pb.StudyResponse{Sid: 1000, Information: "Deng"}
	stream.Send(resp) //服务器流模式发送数据。此处不用关闭发送，因为没找到关闭的方法。
	return nil
}

//双向模式
func (ss *SServer) GetStudy(stream pb.SelfManage_GetStudyServer) error {
	fmt.Println("GetStudy================双向模式================")
	inChan := make(chan *pb.StudyRequest, 1)
	outChan := make(chan struct{}, 1)
	inputChan := make(chan string, 1)
	go func() {
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
				resp := &pb.StudyResponse{Sid: 1000, Information: text}
				stream.Send(resp)
			case <-outChan:
				fmt.Println("server over")
				return
			}
		}
	}()

	for {
		//接收
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("server read done")
			return nil
		}
		select {
		case inChan <- in:
		case <-outChan:
			fmt.Println("server over")
			return nil
		}
	}
	defer func() {
		close(inChan)
		close(outChan)
		close(inputChan)
	}()

	return nil
}

func main() {
	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		grpclog.Fatal("error occur in listen")
		log.Fatal("test")
	}
	s := grpc.NewServer()
	pb.RegisterSelfManageServer(s, &SServer{})
	s.Serve(lis)
}
