package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
	"time"
)

// Message estrutura a mensagem JSON (deve corresponder ao servidor)
type Message struct {
	CMD         string `json:"cmd"`
	Topic       string `json:"topic"`
	Payload     string `json:"payload"`
	ReqID       string `json:"reqId,omitempty"`
	Payload_err string `json:"payload_err,omitempty"`
	FromID      string `json:"fromId,omitempty"`
	Topic_      string `json:"topic_,omitempty"`
}
type ConnectionInfo struct {
	Username string
	Password string
	Host     string
	Port     string
}

// MQ representa o cliente NATS
type MQ struct {
	conn      net.Conn
	reqs      map[string]chan Message
	mu        sync.RWMutex
	url       string
	ID        string
	connected bool
	subs_fun  map[string][]func(message, topic string)
	services  map[string]func(data string, replay func(err, data string))
	Quit      chan struct{}
}

// NewMQ cria uma nova instância do cliente
func NewMQ() *MQ {
	return &MQ{
		url:      "",
		services: make(map[string]func(data string, replay func(err string, data string))),
		reqs:     make(map[string]chan Message),
		subs_fun: map[string][]func(message, topic string){},
		Quit:     make(chan struct{}),
	}
}
func (mc *MQ) parseMQConnection(connectionString string) (*ConnectionInfo, error) {
	// Parse a URL
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %v", err)
	}

	// Verifica se o esquema é "mq"
	if u.Scheme != "mq" {
		return nil, fmt.Errorf("invalid scheme, expected 'mq'")
	}

	// Obtém username e password
	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	// Separa host e port
	host := u.Hostname()
	port := u.Port()

	return &ConnectionInfo{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}, nil
}

// Connect estabelece conexão com o servidor NATS
func (mc *MQ) Connect(url string) error {
	mc.url = url

	auth, err := mc.parseMQConnection(mc.url)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %v", err)
	}

	conn, err := net.Dial("tcp", auth.Host+":"+auth.Port)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %v", err)
	}
	mc.conn = conn
	mc.connected = true
	go mc.readLoop()

	return mc.auth(auth.Username, auth.Password, 5*time.Second)
}
func (mc *MQ) auth(username, password string, timeout time.Duration) error {
	if !mc.connected {
		return fmt.Errorf("cliente não conectado")
	}

	reqId := fmt.Sprintf("%d", time.Now().UnixNano())
	msg := Message{
		CMD:     "AUTH",
		Topic:   username,
		Payload: password,
		ReqID:   reqId,
	}

	err := mc.send(msg)
	if err != nil {
		return err
	}

	// Cria o canal de resposta
	respChan := make(chan Message, 1)

	mc.mu.Lock()
	mc.reqs[reqId] = respChan
	mc.mu.Unlock()

	// Garante que o canal seja removido após a conclusão
	defer func() {
		mc.mu.Lock()
		delete(mc.reqs, reqId)
		mc.mu.Unlock()
	}()

	// Configura o timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-respChan:
		timer.Stop()
		delete(mc.reqs, reqId)
		if resp.Payload_err != "" {
			return fmt.Errorf("%v", resp.Payload_err)
		}
		mc.ID = resp.Payload
		return nil
	case <-timer.C:
		timer.Stop()
		delete(mc.reqs, reqId)
		return fmt.Errorf("timeout após %v", timeout)
	}
}

// Disconnect encerra a conexão com o servidor
func (mc *MQ) Disconnect() {
	close(mc.Quit)
	if mc.conn != nil {
		mc.conn.Close()
	}
	mc.connected = false
}

// Subscribe inscreve-se em um tópico

func (mc *MQ) send(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("erro ao codificar mensagem: %v", err)
	}
	_, err = fmt.Fprintf(mc.conn, "%s\n", data)
	return err
}
func (mc *MQ) Subscribe(topic string, cb func(message, topic string)) error {
	if !mc.connected {
		return fmt.Errorf("cliente não conectado")
	}
	msg := Message{
		CMD:   "SUB",
		Topic: topic,
	}
	err := mc.send(msg)
	if err != nil {
		return err
	}

	mc.subs_fun[topic] = append(mc.subs_fun[topic], cb)
	return nil
}

// Publish envia uma mensagem para um tópico
func (mc *MQ) Publish(topic, payload string) error {
	if !mc.connected {
		return fmt.Errorf("cliente não conectado")
	}
	msg := Message{
		CMD:     "PUB",
		Topic:   topic,
		Payload: payload,
		ReqID:   fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	err := mc.send(msg)
	if err != nil {
		return err
	}

	return nil
}
func (mc *MQ) Service(name string, cb func(data string, replay func(err, data string))) error {
	if !mc.connected {
		return fmt.Errorf("cliente não conectado")
	}
	msg := Message{
		CMD:   "SER",
		Topic: name,
	}
	err := mc.send(msg)
	if err != nil {
		return err
	}
	mc.services[name] = cb
	return nil
}

// Publish envia uma mensagem para um tópico
func (mc *MQ) Request(name, payload string, timeout time.Duration) (string, error) {
	if !mc.connected {
		return "", fmt.Errorf("cliente não conectado")
	}

	reqId := fmt.Sprintf("%d", time.Now().UnixNano())
	msg := Message{
		CMD:     "REQ",
		Topic:   name,
		Payload: payload,
		ReqID:   reqId,
	}

	err := mc.send(msg)
	if err != nil {
		return "", err
	}

	// Cria o canal de resposta
	respChan := make(chan Message, 1)

	mc.mu.Lock()
	mc.reqs[reqId] = respChan
	mc.mu.Unlock()

	// Garante que o canal seja removido após a conclusão
	defer func() {
		mc.mu.Lock()
		delete(mc.reqs, reqId)
		mc.mu.Unlock()
	}()

	// Configura o timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-respChan:
		timer.Stop()
		delete(mc.reqs, reqId)
		if resp.Payload_err != "" {
			return "", fmt.Errorf("%v", resp.Payload_err)
		}
		return resp.Payload, nil
	case <-timer.C:
		timer.Stop()
		delete(mc.reqs, reqId)
		return "", fmt.Errorf("timeout após %v", timeout)
	}
}

// readLoop lê mensagens do servidor
func (mc *MQ) readLoop() {
	scanner := bufio.NewScanner(mc.conn)
	for scanner.Scan() {
		select {
		case <-mc.Quit:
			return
		default:
			rawMsg := scanner.Text()
			var msg Message
			err := json.Unmarshal([]byte(rawMsg), &msg)
			if err != nil {
				log.Printf("Erro ao decodificar mensagem: %v", err)
				continue
			}

			mc.mu.RLock()
			switch msg.CMD {

			case "PUB":
				if funs, ok := mc.subs_fun[msg.Topic]; ok {
					for _, funSub := range funs {
						go funSub(msg.Payload, msg.Topic_)
					}

				}

			case "REQ":

				if funService, ok := mc.services[msg.Topic]; ok {
					go funService(msg.Payload, func(err, data string) {
						mc.send(Message{
							CMD:         "RES",
							FromID:      msg.FromID,
							Payload:     data,
							Payload_err: err,
							Topic:       msg.Topic,
							ReqID:       msg.ReqID,
						})
					})
				}
			case "RES":
				if ch, ok := mc.reqs[msg.ReqID]; ok {
					ch <- msg
				}
			}

			mc.mu.RUnlock()
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Erro ao ler do servidor: %v", err)
		}
	}
}

// Unsubscribe remove uma subscrição
func (mc *MQ) Unsubscribe(topic string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if ch, ok := mc.reqs[topic]; ok {
		close(ch)
		delete(mc.reqs, topic)
	}
}
