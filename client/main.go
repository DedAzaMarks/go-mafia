package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/DedAzaMarks/go-mafia"

	"github.com/DedAzaMarks/go-mafia/pkg/proto/service"
	"google.golang.org/grpc"
)

type client struct {
	stub         service.MafiaClient
	userID       int32
	name         string
	cliInProcess bool
	messages     []service.CommunicationRequest
	mu           sync.RWMutex
}

func run() {
	fmt.Print("Enter your name: ")
	var name string
	_, err := fmt.Scanln(&name)
	if err != nil {
		log.Fatalf("did not scan: %v", err)
	}
	conn, err := grpc.Dial("127.0.0.1:42069", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("can't close connection: %v", err)
		}
	}(conn)
	c := service.NewMafiaClient(conn)
	client := client{
		stub:     c,
		name:     name,
		messages: make([]service.CommunicationRequest, 0),
		mu:       sync.RWMutex{}}
	client.addUser()
	client.messages = append(client.messages, service.CommunicationRequest{
		UserId:   client.userID,
		DataType: service.CommunicationDataType_HANDSHAKE_MESSAGE,
	})
	var wg sync.WaitGroup
	go func(w *sync.WaitGroup) { client.serverCommunication(); w.Done() }(&wg)
	go func(w *sync.WaitGroup) { client.cliCommunication(); w.Done() }(&wg)
	wg.Wait()
}

func (c *client) addUser() {
	response, err := c.stub.AddUser(context.Background(), &service.AddUserRequest{Name: c.name})
	if err != nil {
		log.Fatalf("failed to recv msg")
	}
	val, _ := strconv.Atoi(response.Message)
	c.userID = int32(val)
}

func (c *client) serverCommunication() {
	channel, err := c.stub.InitCommunicationChannel(context.Background())
	if err != nil {
		log.Fatalf("can't init communication: %v", err)
	}
	send := func() {
		for {
			for message := range c.messages {
				err := channel.SendMsg(message)
				if err != nil {
					log.Fatalf("can't send message: %v", err)
				}
			}
		}
	}
	recv := func() {
		var in service.Response
		for {
			err := channel.RecvMsg(&in)
			if err != nil {
				log.Fatalf("can't reciev respond: %v", err)
			}
		}
	}

	go send()
	go recv()
}

func (c *client) cliCommunication() {
	c.cliInProcess = false
	for {
		if !c.cliInProcess {
			fmt.Print("> ")
			var command string
			_, _ = fmt.Scanf("%s", &command)
			c.executeCommand(command)
			c.cliInProcess = false
		}
	}
}

func (c *client) executeCommand(command string) {
	if len(command) == 0 {
		return
	}
	commandStr := strings.Split(command, " ")
	args := strings.Join(commandStr[1:], " ")
	switch commandStr[0] {
	case mafia.GetUsers:
		response, err := c.stub.GetUsers(context.Background(), &service.Empty{})
		if err != nil {
			log.Fatalf("can't get users: %v", err)
		}
		for id, name := range response.Users {
			fmt.Printf("id: %d | name %s", id, name)
		}
	case mafia.Broadcast:
		message := args
		c.messages = append(c.messages, service.CommunicationRequest{
			UserId:   c.userID,
			Message:  message,
			DataType: service.CommunicationDataType_BROADCAST_MESSAGE,
		})
	case mafia.VoteFinishDay:
		response, err := c.stub.VoteFinishDay(context.Background(), &service.BaseUserRequest{UserId: c.userID})
		if err != nil {
			log.Fatalf("can't vote")
		}
		fmt.Println(response.String())
	case mafia.Decision:
		c.messages = append(c.messages, service.CommunicationRequest{
			UserId:   c.userID,
			Message:  args,
			DataType: service.CommunicationDataType_DECISION_MESSAGE,
		})
	case mafia.Accuse:
		val, err := strconv.Atoi(args)
		if err != nil {
			log.Fatalf("number expected, got %s", err)
		}
		response, err := c.stub.AccuseUser(context.Background(), &service.AccuseUserRequest{
			AccusingUserId: c.userID,
			AccusedUserId:  int32(val),
		})
		fmt.Println(response.String())
	default:
		fmt.Println("unknown command")
	}
	fmt.Printf("Executed %s\n", commandStr)
}

func main() {
	run()
}
