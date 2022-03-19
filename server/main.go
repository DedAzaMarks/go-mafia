package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/DedAzaMarks/go-mafia"

	"github.com/DedAzaMarks/go-mafia/pkg/proto/service"

	"google.golang.org/grpc"
)

type server struct {
	service.UnimplementedMafiaServer
	users          map[int32]*service.User
	roles          map[int32]int
	userMessages   map[int32][]service.CommunicationResponse
	accuseVotes    map[int32]int32
	mafiaVotes     map[int32]int32
	detectiveVotes map[int32]int32
	finishDayVotes map[int32]bool
	started        bool
	finished       bool
	dayNum         uint32
	dayTime        int
}

func (s *server) AddUser(_ context.Context, req *service.AddUserRequest) (*service.Response, error) {
	userID := int32(len(s.users))
	name := req.Name
	for _, ok := s.users[userID]; ok; {
		userID++
	}
	for id := range s.users {
		s.userMessages[id] = append(s.userMessages[id], service.CommunicationResponse{
			Message: fmt.Sprintf("User %s has been added", name),
			Author:  "Server"})
	}
	user := service.User{UserId: userID, Name: name}
	s.users[userID] = &user
	s.roles[userID] = mafia.RoleNotAssigned
	s.userMessages[userID] = make([]service.CommunicationResponse, 0)

	return &service.Response{
		Status: service.StatusCode_CREATED,
		Data: map[string]string{
			strconv.Itoa(int(userID)): strconv.Itoa(int(userID)),
		}}, nil
}

func (s *server) DeleteUser(_ context.Context, req *service.BaseUserRequest) (*service.Response, error) {
	if _, ok := s.users[req.UserId]; ok {
		delete(s.users, req.UserId)
		return &service.Response{
			Status:  service.StatusCode_OK,
			Message: fmt.Sprintf("No user with ID: %d", req.UserId),
		}, nil
	}
	return &service.Response{
		Status:  service.StatusCode_NOT_FOUND,
		Message: fmt.Sprintf("Deleted user ID: %d", req.UserId),
	}, nil
}

func (s *server) GetUsers(context.Context, *service.Empty) (*service.GetUsersResponse, error) {
	var res strings.Builder
	for userID := range s.users {
		res.WriteString(
			fmt.Sprintf("username: %s, userID: %d, alive: %t",
				s.users[userID].Name,
				s.users[userID].UserId,
				s.roles[userID] != mafia.RoleGhost))
	}
	return &service.GetUsersResponse{
		Status: service.StatusCode_OK,
		Users:  s.users,
	}, nil
}

func (s *server) AccuseUser(_ context.Context, req *service.AccuseUserRequest) (*service.Response, error) {
	if s.dayNum <= 1 {
		return &service.Response{
			Status:  service.StatusCode_BAD_REQUEST,
			Message: "You can't vote on the first day",
		}, nil
	}
	if s.roles[req.AccusedUserId] == mafia.RoleGhost {
		return &service.Response{
			Status:  service.StatusCode_BAD_REQUEST,
			Message: fmt.Sprintf("User %d is already a ghost", req.AccusedUserId),
		}, nil
	}
	_, ok1 := s.users[req.AccusingUserId]
	_, ok2 := s.users[req.AccusedUserId]
	if ok1 && ok2 {
		s.accuseVotes[req.AccusingUserId] = req.AccusedUserId
		return &service.Response{
			Status:  service.StatusCode_OK,
			Message: "OK",
		}, nil
	}
	return &service.Response{
		Status:  service.StatusCode_BAD_REQUEST,
		Message: "Bad user IDs",
	}, nil
}

func (s *server) VoteFinishDay(_ context.Context, req *service.BaseUserRequest) (*service.Response, error) {
	if _, ok := s.users[req.UserId]; !ok {
		return &service.Response{
			Status:  service.StatusCode_BAD_REQUEST,
			Message: "Bad user ID",
		}, nil
	}
	s.finishDayVotes[req.UserId] = true
	for userID := range s.users {
		if userID != req.UserId {
			s.userMessages[userID] = append(s.userMessages[userID], service.CommunicationResponse{
				Message: "User votes to finish day",
				Author:  "Server"})
		}
	}
	return &service.Response{
		Status:  service.StatusCode_OK,
		Message: "Your votes has been counted",
	}, nil
}

func (s *server) InitCommunicationChannel(channel service.Mafia_InitCommunicationChannelServer) error {
	for {
		var in service.CommunicationRequest
		if err := channel.RecvMsg(&in); err != nil {
			return fmt.Errorf("failed to reciev request: %w", err)
		}
		switch in.DataType {
		case service.CommunicationDataType_HANDSHAKE_MESSAGE:
			log.Printf("Handshake with user %d\n", in.UserId)
			if err := channel.SendMsg(&service.CommunicationResponse{
				Message: "You joined Mafia game",
				Author:  "Server",
			}); err != nil {
				return fmt.Errorf("failed to send respond: %w", err)
			}
		case service.CommunicationDataType_BROADCAST_MESSAGE:
			log.Printf("Recieved broadcast message from %d", in.UserId)
			toRole := -1
			if s.dayTime == mafia.DayTimeNight {
				switch s.roles[in.UserId] {
				case mafia.RoleMafia:
					toRole = mafia.RoleMafia
				case mafia.RoleDetective:
					toRole = mafia.RoleDetective
				default:
					if err := channel.SendMsg(&service.CommunicationResponse{
						Message: "You can't message during night time",
						Author:  "Server",
					}); err != nil {
						return fmt.Errorf("failed to send respond: %w", err)
					}
				}
			}
			for userID := range s.users {
				if in.UserId == userID {
					continue
				}
				switch toRole {
				case mafia.RoleMafia, mafia.RoleDetective:
					if s.roles[userID] == toRole {
						s.userMessages[userID] = append(
							s.userMessages[userID],
							service.CommunicationResponse{
								Message: in.Message,
								Author:  s.users[in.UserId].Name,
							})
					}
				default:
					s.userMessages[userID] = append(
						s.userMessages[userID],
						service.CommunicationResponse{
							Message: in.Message,
							Author:  s.users[in.UserId].Name,
						})
				}
			}
		case service.CommunicationDataType_DECISION_MESSAGE:
			log.Printf("Received mafia message from %d", in.UserId)
			log.Printf("Decision request: %s", in.String())
			log.Printf("Mafia votes: %v", s.mafiaVotes)
			log.Printf("Detective votes: %v", s.detectiveVotes)
			t, err := strconv.Atoi(in.Message)
			target := int32(t)
			if err != nil {
				if err := channel.SendMsg(&service.CommunicationResponse{
					Message: "Can't parse your message",
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("can't send respond: %w", err)
				}
			}

			if in.UserId != mafia.RoleMafia && in.UserId != mafia.RoleDetective {
				if err := channel.SendMsg(&service.CommunicationResponse{
					Message: "You are not mafia or detective",
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			}
			var votes map[int32]int32
			if _, ok := s.users[target]; !ok {
				if err := channel.SendMsg(&service.CommunicationResponse{
					Message: fmt.Sprintf("No user with this ID: %d", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else if s.roles[target] == mafia.RoleGhost {
				if err := channel.SendMsg(&service.CommunicationResponse{
					Message: fmt.Sprintf("User with %d is already dead", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else if s.roles[in.UserId] == s.roles[target] {
				if err := channel.SendMsg(&service.CommunicationResponse{
					Message: fmt.Sprintf("User with %d has the same role with you", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else {
				_, isMafiaVote := s.mafiaVotes[in.UserId]
				_, isDetectiveVote := s.detectiveVotes[in.UserId]
				if (s.roles[in.UserId] == mafia.RoleMafia && isMafiaVote) || (s.roles[in.UserId] == mafia.RoleDetective && isDetectiveVote) {
					if err := channel.SendMsg(&service.CommunicationResponse{
						Message: "You've already voted",
						Author:  "Server",
					}); err != nil {
						return fmt.Errorf("failed to send respond: %w", err)
					}
				} else {
					switch s.roles[in.UserId] {
					case mafia.RoleMafia:
						votes = s.mafiaVotes
					case mafia.RoleDetective:
						votes = s.detectiveVotes
					default:
						return fmt.Errorf("invalid role")
					}
					votes[in.UserId] = target
					for userID := range s.users {
						if userID != in.UserId && s.roles[userID] == s.roles[in.UserId] {
							if err := channel.SendMsg(&service.CommunicationResponse{
								Message: fmt.Sprintf("I vote for %d ", target),
								Author:  s.users[in.UserId].Name,
							}); err != nil {
								return fmt.Errorf("failed to send respond: %w", err)
							}
						}
					}
					if err := channel.SendMsg(&service.CommunicationResponse{
						Message: fmt.Sprintf("You voted for user %d", target),
						Author:  "Server",
					}); err != nil {
						return fmt.Errorf("failed to send respond: %w", err)
					}
				}
			}
		case service.CommunicationDataType_EMPTY_MESSAGE:
			continue
		default:
		}
		if _, ok := s.userMessages[in.UserId]; ok && len(s.userMessages[in.UserId]) != 0 {
			if err := channel.SendMsg(
				&service.CommunicationResponse{
					Message: s.userMessages[in.UserId][0].Message,
					Author:  s.userMessages[in.UserId][0].Author,
				}); err != nil {
				return fmt.Errorf("can't send respond: %w", err)
			}
			s.userMessages[in.UserId] = s.userMessages[in.UserId][1:]
		} else {
			if err := channel.SendMsg(&service.CommunicationResponse{}); err != nil {
				return fmt.Errorf("can't send respond: %w", err)
			}
		}
	}
}

func main() {
	s := grpc.NewServer()
	service.RegisterMafiaServer(s, &server{
		users:          make(map[int32]*service.User),
		roles:          make(map[int32]int),
		userMessages:   make(map[int32][]service.CommunicationResponse),
		accuseVotes:    make(map[int32]int32),
		mafiaVotes:     make(map[int32]int32),
		detectiveVotes: make(map[int32]int32),
		finishDayVotes: make(map[int32]bool),
		started:        false,
		finished:       false,
		dayNum:         0,
		dayTime:        mafia.DayTimeDay,
	})
	listener, err := net.Listen("tcp", "0.0.0.0::42069")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
