package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/DedAzaMarks/go-mafia"

	"github.com/DedAzaMarks/go-mafia/pkg/proto/service"

	"google.golang.org/grpc"
)

func getRole(i int) string {
	switch i {
	case mafia.RoleInnocent:
		return "Innocent"
	case mafia.RoleMafia:
		return "Mafia"
	case mafia.RoleDetective:
		return "Detective"
	case mafia.RoleGhost:
		return "Ghost"
	case mafia.RoleNotAssigned:
		return ""
	default:
		log.Fatalf("undefined role")
	}
	return "" // unreachable
}

type server struct {
	service.UnimplementedMafiaServer
	users          map[int32]*service.User
	roles          map[int32]int
	userMessages   map[int32][]service.Response
	accuseVotes    map[int32]int32
	mafiaVotes     map[int32]int32
	detectiveVotes map[int32]int32
	finishDayVotes map[int32]bool
	gameStarted    bool
	gameFinished   bool
	dayNum         uint32
	dayTime        int
	mu             sync.RWMutex
}

func (s *server) AddUser(_ context.Context, req *service.AddUserRequest) (*service.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userID := int32(len(s.users))
	name := req.Name
	for _, ok := s.users[userID]; ok; {
		userID++
	}
	for id := range s.users {
		s.userMessages[id] = append(s.userMessages[id], service.Response{
			Status:  service.StatusCode_OK,
			Message: fmt.Sprintf("User %s has been added", name),
			Author:  "Server"})
	}
	user := service.User{UserId: userID, Name: name}
	s.users[userID] = &user
	s.roles[userID] = mafia.RoleNotAssigned
	s.userMessages[userID] = make([]service.Response, 0)

	return &service.Response{
		Status:  service.StatusCode_CREATED,
		Message: strconv.Itoa(int(userID)),
		Author:  "Server",
	}, nil
}

func (s *server) DeleteUser(_ context.Context, req *service.BaseUserRequest) (*service.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[req.UserId]; ok {
		delete(s.users, req.UserId)
		return &service.Response{
			Status:  service.StatusCode_OK,
			Message: fmt.Sprintf("No user with ID: %d", req.UserId),
			Author:  "Server",
		}, nil
	}
	return &service.Response{
		Status:  service.StatusCode_NOT_FOUND,
		Message: fmt.Sprintf("Deleted user ID: %d", req.UserId),
		Author:  "Server",
	}, nil
}

func (s *server) isAlive(userID int32) bool {
	return s.roles[userID] != mafia.RoleGhost
}

func (s *server) GetUsers(context.Context, *service.Empty) (*service.GetUsersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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
			Author:  "Server",
		}, nil
	}
	if s.roles[req.AccusedUserId] == mafia.RoleGhost {
		return &service.Response{
			Status:  service.StatusCode_BAD_REQUEST,
			Message: fmt.Sprintf("User %d is already a ghost", req.AccusedUserId),
			Author:  "Server",
		}, nil
	}
	_, ok1 := s.users[req.AccusingUserId]
	_, ok2 := s.users[req.AccusedUserId]
	if ok1 && ok2 {
		s.accuseVotes[req.AccusingUserId] = req.AccusedUserId
		return &service.Response{
			Status:  service.StatusCode_OK,
			Message: "OK",
			Author:  "Server",
		}, nil
	}
	return &service.Response{
		Status:  service.StatusCode_BAD_REQUEST,
		Message: "Bad user IDs",
	}, nil
}

func (s *server) VoteFinishDay(_ context.Context, req *service.BaseUserRequest) (*service.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[req.UserId]; !ok {
		return &service.Response{
			Status:  service.StatusCode_BAD_REQUEST,
			Message: "Bad user ID",
			Author:  "Server",
		}, nil
	}
	s.finishDayVotes[req.UserId] = true
	for userID := range s.users {
		if userID != req.UserId {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "User votes to finish day",
				Author:  "Server",
			})
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
			if err := channel.SendMsg(&service.Response{
				Status:  service.StatusCode_OK,
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
					if err := channel.SendMsg(&service.Response{
						Status:  service.StatusCode_FORBIDDEN,
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
							service.Response{
								Status:  service.StatusCode_OK,
								Message: in.Message,
								Author:  s.users[in.UserId].Name,
							})
					}
				default:
					s.userMessages[userID] = append(
						s.userMessages[userID],
						service.Response{
							Status:  service.StatusCode_OK,
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
				if err := channel.SendMsg(&service.Response{
					Status:  service.StatusCode_BAD_REQUEST,
					Message: "Can't parse your message",
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("can't send respond: %w", err)
				}
			}

			if in.UserId != mafia.RoleMafia && in.UserId != mafia.RoleDetective {
				if err := channel.SendMsg(&service.Response{
					Status:  service.StatusCode_FORBIDDEN,
					Message: "You are not mafia or detective",
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			}
			var votes map[int32]int32
			if _, ok := s.users[target]; !ok {
				if err := channel.SendMsg(&service.Response{
					Status:  service.StatusCode_NOT_FOUND,
					Message: fmt.Sprintf("No user with this ID: %d", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else if s.roles[target] == mafia.RoleGhost {
				if err := channel.SendMsg(&service.Response{
					Status:  service.StatusCode_BAD_REQUEST,
					Message: fmt.Sprintf("User with %d is already dead", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else if s.roles[in.UserId] == s.roles[target] {
				if err := channel.SendMsg(&service.Response{
					Status:  service.StatusCode_BAD_REQUEST,
					Message: fmt.Sprintf("User with %d has the same role with you", target),
					Author:  "Server",
				}); err != nil {
					return fmt.Errorf("failed to send respond: %w", err)
				}
			} else {
				_, isMafiaVote := s.mafiaVotes[in.UserId]
				_, isDetectiveVote := s.detectiveVotes[in.UserId]
				if (s.roles[in.UserId] == mafia.RoleMafia && isMafiaVote) || (s.roles[in.UserId] == mafia.RoleDetective && isDetectiveVote) {
					if err := channel.SendMsg(&service.Response{
						Status:  service.StatusCode_BAD_REQUEST,
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
						return fmt.Errorf("undefined role")
					}
					votes[in.UserId] = target
					for userID := range s.users {
						if userID != in.UserId && s.roles[userID] == s.roles[in.UserId] {
							if err := channel.SendMsg(&service.Response{
								Status:  service.StatusCode_OK,
								Message: fmt.Sprintf("I vote for %d ", target),
								Author:  s.users[in.UserId].Name,
							}); err != nil {
								return fmt.Errorf("failed to send respond: %w", err)
							}
						}
					}
					if err := channel.SendMsg(&service.Response{
						Status:  service.StatusCode_OK,
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
				&service.Response{
					Status:  service.StatusCode_OK,
					Message: s.userMessages[in.UserId][0].Message,
					Author:  s.userMessages[in.UserId][0].Author,
				}); err != nil {
				return fmt.Errorf("can't send respond: %w", err)
			}
			s.userMessages[in.UserId] = s.userMessages[in.UserId][1:]
		} else {
			if err := channel.SendMsg(&service.Response{}); err != nil {
				return fmt.Errorf("can't send respond: %w", err)
			}
		}
	}
}

func (s *server) isMafiaWinner() bool {
	aliveMafia, aliveOther := 0, 0
	for userID := range s.users {
		if s.isAlive(userID) {
			if s.roles[userID] == mafia.RoleMafia {
				aliveMafia++
			} else {
				aliveOther++
			}
		}
	}
	return aliveMafia >= aliveOther
}

func (s *server) run(wg *sync.WaitGroup) {
	defer wg.Done()
	checkVotes := func(votes map[int32]bool) bool {
		for _, vote := range votes {
			if !vote {
				return false
			}
		}
		return true
	}
	checkNightVotes := func(votes map[int32]int32) bool {
		for userID := range votes {
			if _, ok := votes[userID]; !ok {
				return false
			}
		}
		return true
	}
	for {
		if !s.gameStarted && len(s.users) >= mafia.MinUsersNum {
			s.startGame()
		}
		if !s.gameStarted {
			continue
		}
		if checkVotes(s.finishDayVotes) && s.dayTime == mafia.DayTimeDay {
			s.startNight()
		} else if s.dayTime == mafia.DayTimeNight && checkNightVotes(s.mafiaVotes) && checkNightVotes(s.detectiveVotes) {
			log.Printf("everybody voted, finish night")
			s.startDay()
		} else if s.isMafiaWinner() && !s.gameFinished {
			s.gameFinished = true
			for userID := range s.users {
				s.userMessages[userID] = append(s.userMessages[userID], service.Response{
					Status:  service.StatusCode_OK,
					Message: "Mafia win!",
					Author:  "Server"})
			}
		}
	}
}

func (s *server) startDay() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("start day")
	if len(s.users) < mafia.MinUsersNum {
		log.Fatalf("not enough players")
	}
	if s.dayNum != 0 {
		s.executeMafiaDecision()
		s.executeDetectiveDecision()
	}
	s.dayTime = mafia.DayTimeDay
	s.dayNum++
	s.accuseVotes = make(map[int32]int32)
	s.finishDayVotes = make(map[int32]bool)
}

func (s *server) startNight() {
	log.Printf("start night")
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dayNum != 1 {
		s.killAccusedUser()
	}
	s.mafiaVotes, s.detectiveVotes = make(map[int32]int32), make(map[int32]int32)
	s.dayTime = mafia.DayTimeNight
	s.finishDayVotes = make(map[int32]bool)
	for userID := range s.users {
		s.finishDayVotes[userID] = false
	}
	for userID := range s.users {
		s.userMessages[userID] = append(s.userMessages[userID], service.Response{
			Status:  service.StatusCode_OK,
			Message: "Night started",
			Author:  "Server",
		})
		switch s.roles[userID] {
		case mafia.RoleMafia:
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "Choose who to kill",
				Author:  "Server",
			})
		case mafia.RoleDetective:
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "Choose who to check",
				Author:  "Server",
			})
		}
	}
}

func (s *server) assignRoles() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.users) < mafia.MinUsersNum {
		log.Fatalf("not enough players")
	}
	var userIDs []int32
	for userID := range s.users {
		userIDs = append(userIDs, userID)
	}
	rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })
	for userID := range userIDs[:mafia.MafiaNum] {
		s.roles[int32(userID)] = mafia.RoleMafia
	}
	s.roles[userIDs[mafia.MafiaNum]] = mafia.RoleDetective
	for userID := range userIDs[:mafia.MafiaNum] {
		s.roles[int32(userID)] = mafia.RoleInnocent
	}
}

func (s *server) startGame() {
	log.Printf("start game")
	s.gameStarted = true
	s.assignRoles()
	s.startDay()
	log.Printf("game started")
	for usedID := range s.users {
		s.userMessages[usedID] = append(s.userMessages[usedID], service.Response{
			Status:  service.StatusCode_OK,
			Message: fmt.Sprintf("The game just started. You are %s!", getRole(s.roles[usedID])),
			Author:  "Server"})
	}
}

func (s *server) killAccusedUser() {
	votesCounter := make(map[int32]int32)
	for _, localTarget := range s.accuseVotes {
		votesCounter[localTarget]++
	}
	var target int32 = 0
	for targetID, votesSum := range votesCounter {
		if votesSum > target {
			target = targetID
		}
	}
	s.roles[target] = mafia.RoleGhost
	for userID := range s.users {
		if userID == target {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "Majority voted for you, you are a ghost now",
				Author:  "Server",
			})
		} else {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: fmt.Sprintf("User %d was killed by votes", target),
				Author:  "Server",
			})
		}
	}
}

func (s *server) executeMafiaDecision() {
	votesCounter := make(map[int32]int32)
	for _, localTarget := range s.mafiaVotes {
		votesCounter[localTarget]++
	}
	var target int32 = 0
	for targetID, votesSum := range votesCounter {
		if votesSum > target {
			target = targetID
		}
	}
	s.roles[target] = mafia.RoleGhost
	for userID := range s.users {
		if userID == target {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "Mafia killed you",
				Author:  "Server",
			})
		} else {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: fmt.Sprintf("Mafia killed user %d", target),
				Author:  "Server",
			})
		}
	}
}

func (s *server) executeDetectiveDecision() {
	votesCounter := make(map[int32]int32)
	for _, localTarget := range s.detectiveVotes {
		votesCounter[localTarget]++
	}
	var target int32 = 0
	for targetID, votesSum := range votesCounter {
		if votesSum > target {
			target = targetID
		}
	}
	if s.roles[target] == mafia.RoleMafia {
		s.roles[target] = mafia.RoleGhost
		for userID := range s.users {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: fmt.Sprintf("Detective found mafia: user %d", target),
				Author:  "Server",
			})
		}
	} else {
		for userID := range s.users {
			s.userMessages[userID] = append(s.userMessages[userID], service.Response{
				Status:  service.StatusCode_OK,
				Message: "Detective didnt find mafia",
				Author:  "Server",
			})
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	s := grpc.NewServer()
	server := server{
		users:          make(map[int32]*service.User),
		roles:          make(map[int32]int),
		userMessages:   make(map[int32][]service.Response),
		accuseVotes:    make(map[int32]int32),
		mafiaVotes:     make(map[int32]int32),
		detectiveVotes: make(map[int32]int32),
		finishDayVotes: make(map[int32]bool),
		gameStarted:    false,
		gameFinished:   false,
		dayNum:         0,
		dayTime:        mafia.DayTimeDay,
		mu:             sync.RWMutex{},
	}
	service.RegisterMafiaServer(s, &server)
	listener, err := net.Listen("tcp", "0.0.0.0:42069")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go server.run(&wg)
	go func(lis net.Listener, wg *sync.WaitGroup) {
		err := s.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		wg.Done()
	}(listener, &wg)
	wg.Wait()
}
