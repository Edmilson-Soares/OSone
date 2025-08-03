package mq

// Subscribe adiciona um novo assinante para um tópico

func (mq *MQ) handleService(clientId, virtual, topic string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.services[virtual] == nil {
		mq.services[virtual] = make(map[string]string)
	}
	mq.services[virtual][topic] = clientId
}

func (mq *MQ) Service(virtual, name string, cb func(data string, replay func(err, data string))) error {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	if mq.services_fun[virtual] == nil {
		mq.services_fun[virtual] = make(map[string]func(data string, replay func(err string, data string)))
	}
	if mq.services[virtual] == nil {
		mq.services[virtual] = make(map[string]string)
	}
	mq.services_fun[virtual][name] = cb
	mq.services[virtual][name] = "mybroker"
	return nil
}
