package api

import (
	"github.com/mellojp/chatli/data"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func getAPIURL() string {
	val := os.Getenv("API_URL")
	if val == "" {
		return "http://localhost:8080"
	}
	return val
}

// Login realiza a autenticação e retorna a sessão com o token JWT
func Login(username, password string) (*data.Session, error) {
	msg := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(msg)

	resp, err := http.Post(getAPIURL()+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login falhou: status %d", resp.StatusCode)
	}

	var res map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	session := &data.Session{
		Token:       res["token"],
		Username:    username,
		UserId:      res["user_id"],
		JoinedRooms: []data.Room{},
	}
	return session, nil
}

// Register cria um novo usuário
func Register(username, password string) error {
	msg := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(msg)

	resp, err := http.Post(getAPIURL()+"/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("registro falhou: status %d", resp.StatusCode)
	}
	return nil
}

// CreateRoom cria uma nova sala enviando o nome desejado
func CreateRoom(s data.Session, roomName string) (*data.Room, error) {
	reqUrl := getAPIURL() + "/rooms/create"
	payload := map[string]string{"room_name": roomName}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+s.Token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("erro ao criar sala: status %d", resp.StatusCode)
	}

	var res data.Room
	json.NewDecoder(resp.Body).Decode(&res)
	return &res, nil
}

func JoinRoom(s data.Session, roomId string) error {
	reqUrl := getAPIURL() + "/rooms/join"
	payload := map[string]string{"room_id": roomId}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+s.Token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("erro ao entrar na sala: status %d", resp.StatusCode)
	}
	return nil
}

// GetUserRooms busca as salas que o usuário já entrou
func GetUserRooms(s data.Session) ([]data.Room, error) {
	reqUrl := getAPIURL() + "/rooms"
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("Authorization", "Bearer "+s.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro ao buscar salas: status %d", resp.StatusCode)
	}

	var rooms []data.Room
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func LoadChatMessages(s data.Session, roomId string) ([]data.Message, error) {
	reqUrl := fmt.Sprintf("%s/rooms/history?room_id=%s", getAPIURL(), roomId)
	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("Authorization", "Bearer "+s.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro na requisição: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var messages []data.Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []data.Message{}
	}
	return messages, nil
}
