import  net from 'net';
import  EventEmitter from 'events';
import { URL } from 'url';

class Message {
    constructor({ cmd, topic, payload, reqId, payload_err, fromId, topic_ }) {
        this.CMD = cmd;
        this.Topic = topic;
        this.Payload = payload;
        this.ReqID = reqId;
        this.Payload_err = payload_err;
        this.FromID = fromId;
        this.Topic_ = topic_;
    }
}

class ConnectionInfo {
    constructor(username, password, host, port) {
        this.Username = username;
        this.Password = password;
        this.Host = host;
        this.Port = port;
    }
}

class MQ extends EventEmitter {
    constructor() {
        super();
        this.conn = null;
        this.reqs = new Map();
        this.url = '';
        this.ID = '';
        this.connected = false;
        this.subs_fun = new Map();
        this.services = new Map();
        this.Quit = false;
    }

    parseMQConnection(connectionString) {
        try {
            const u = new URL(connectionString);
            
            if (u.protocol !== 'mq:') {
                throw new Error("Invalid scheme, expected 'mq'");
            }

            const username = u.username || '';
            const password = u.password || '';
            const host = u.hostname;
            const port = u.port;

            return new ConnectionInfo(username, password, host, port);
        } catch (err) {
            throw new Error(`Failed to parse connection string: ${err.message}`);
        }
    }

    async connect(url) {
        this.url = url;

        const auth = this.parseMQConnection(this.url);
        
        return new Promise((resolve, reject) => {
            this.conn = net.createConnection({
                host: auth.Host,
                port: auth.Port
            }, () => {
                this.connected = true;
                this.readLoop();
                this.auth(auth.Username, auth.Password, 5000)
                    .then(resolve)
                    .catch(reject);
            });

            this.conn.on('error', (err) => {
                reject(new Error(`Error connecting to server: ${err.message}`));
            });
        });
    }

    async auth(username, password, timeout) {
        if (!this.connected) {
            throw new Error("Client not connected");
        }

        const reqId = Date.now().toString();
        const msg = new Message({
            cmd: 'AUTH',
            topic: username,
            payload: password,
            reqId: reqId
        });

        await this.send(msg);

        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.reqs.delete(reqId);
                reject(new Error(`Timeout after ${timeout}ms`));
            }, timeout);

            this.reqs.set(reqId, (resp) => {
                clearTimeout(timer);
                this.reqs.delete(reqId);
                if (resp.Payload_err) {
                    reject(new Error(resp.Payload_err));
                } else {
                    this.ID = resp.Payload;
                    resolve();
                }
            });
        });
    }

    disconnect() {
        this.Quit = true;
        if (this.conn) {
            this.conn.end();
        }
        this.connected = false;
    }

    async send(msg) {
        return new Promise((resolve, reject) => {
            try {
                const data = JSON.stringify(msg) + '\n';
                this.conn.write(data, (err) => {
                    if (err) {
                        reject(new Error(`Error sending message: ${err.message}`));
                    } else {
                        resolve();
                    }
                });
            } catch (err) {
                reject(new Error(`Error encoding message: ${err.message}`));
            }
        });
    }

    async subscribe(topic, cb) {
        if (!this.connected) {
            throw new Error("Client not connected");
        }

        const msg = new Message({
            cmd: 'SUB',
            topic: topic
        });

        await this.send(msg);

        if (!this.subs_fun.has(topic)) {
            this.subs_fun.set(topic, []);
        }
        this.subs_fun.get(topic).push(cb);
    }

    async publish(topic, payload) {
        let input=payload
        if (!this.connected) {
            throw new Error("Client not connected");
        }

        try {
          input=  JSON.stringify(payload)
        } catch (error) {
            
        }

        const msg = new Message({
            cmd: 'PUB',
            topic: topic,
            payload: input,
            reqId: Date.now().toString()
        });

        await this.send(msg);
    }

    async service(name, cb) {
        if (!this.connected) {
            throw new Error("Client not connected");
        }

        const msg = new Message({
            cmd: 'SER',
            topic: name
        });

        await this.send(msg);
        this.services.set(name, cb);
    }

    async request(name, payload, timeout = 5000) {
                let input=payload
        if (!this.connected) {
            throw new Error("Client not connected");
        }
      try {
          input=  JSON.stringify(payload)
        } catch (error) {
            
        }
        const reqId = Date.now().toString();
        const msg = new Message({
            cmd: 'REQ',
            topic: name,
            payload: input,
            reqId: reqId
        });

        await this.send(msg);

        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.reqs.delete(reqId);
                reject(new Error(`Timeout after ${timeout}ms`));
            }, timeout);

            this.reqs.set(reqId, (resp) => {
                clearTimeout(timer);
                this.reqs.delete(reqId);
                if (resp.Payload_err) {
                    reject(new Error(resp.Payload_err));
                } else {
                    resolve(resp.Payload);
                }
            });
        });
    }

    readLoop() {
        let buffer = '';
        
        this.conn.on('data', (data) => {
            if (this.Quit) return;

            buffer += data.toString();
            const messages = buffer.split('\n');
            
            // The last element might be incomplete, so we keep it in the buffer
            buffer = messages.pop() || '';
            
            for (const rawMsg of messages) {
                if (!rawMsg.trim()) continue;
                
                try {
                    const msg = new Message(JSON.parse(rawMsg));
                    
                    switch (msg.CMD) {
                        case 'PUB':
                            if (this.subs_fun.has(msg.Topic)) {
                                const funs = this.subs_fun.get(msg.Topic);
                                for (const funSub of funs) {
                                    let payload = msg.Payload;
                                    try {
                                        payload=JSON.parse( msg.Payload)
                                    } catch (error) {
                                        
                                    }
                                    setImmediate(() => funSub(payload, msg.Topic_));
                                }
                            }
                            break;
                            
                        case 'REQ':
                            if (this.services.has(msg.Topic)) {
                                const funService = this.services.get(msg.Topic);
                                setImmediate(() => {
                                    funService(msg.Payload, (err, data) => {
                                              let output =data;
                                                try {
                                                    output=JSON.stringify(data)
                                                } catch (error) {
                                                    
                                                }
                                        this.send(new Message({
                                            cmd: 'RES',
                                            fromId: msg.FromID,
                                            payload: output,
                                            payload_err: err,
                                            topic: msg.Topic,
                                            reqId: msg.ReqID
                                        })).catch(console.error);
                                    });
                                });
                            }
                            break;
                            
                        case 'RES':
                            if (this.reqs.has(msg.ReqID)) {
                                const callback = this.reqs.get(msg.ReqID);
                                 try {
                                  msg.Payload=JSON.parse(msg.Payload)
                                } catch (error) {
                                                    
                                }
                                callback(msg);
                            }
                            break;
                    }
                } catch (err) {
                    console.error(`Error decoding message: ${err.message}`);
                }
            }
        });

        this.conn.on('error', (err) => {
            if (!this.Quit) {
                console.error(`Error reading from server: ${err.message}`);
            }
        });

        this.conn.on('close', () => {
            if (!this.Quit) {
                this.connected = false;
                this.emit('disconnected');
            }
        });
    }

    unsubscribe(topic) {
        this.subs_fun.delete(topic);
    }
}

export { MQ, Message };