import net from 'net';
import { EventEmitter } from 'events';
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
    constructor(options = {}) {
        super();
        this.conn = null;
        this.reqs = new Map();
        this.url = '';
        this.ID = '';
        this.connected = false;
        this.subs_fun = new Map();
        this.services = new Map();
        this.Quit = false;

        // Configuration options
        this.options = {
            autoReconnect: true,
            reconnectDelay: 1000,
            maxReconnectAttempts: Infinity,
            connectTimeout: 5000,
            ...options
        };

        this.reconnectAttempts = 0;
        this.reconnectTimer = null;
        this.pendingMessages = [];
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
        if (this.connected) {
            return Promise.resolve();
        }

        this.url = url;
        this.Quit = false;
        this.reconnectAttempts = 0;

        return this._attemptConnect();
    }

    async _attemptConnect() {
        const auth = this.parseMQConnection(this.url);

        return new Promise((resolve, reject) => {
            // Connection timeout
            const connectTimer = setTimeout(() => {
                cleanup();
                reject(new Error(`Connection timeout after ${this.options.connectTimeout}ms`));
            }, this.options.connectTimeout);

            const cleanup = () => {
                clearTimeout(connectTimer);
                this.conn?.removeAllListeners();
            };

            this.conn = net.createConnection({
                host: auth.Host,
                port: auth.Port
            }, () => {
                cleanup();
                this.connected = true;
                this.reconnectAttempts = 0;
                this.emit('connect');
                this.readLoop();

                this.auth(auth.Username, auth.Password, 5000)
                    .then(() => {
                        this.emit('ready');
                        this._flushPendingMessages();
                        resolve();
                    })
                    .catch(err => {
                        this.conn.end();
                        reject(err);
                    });
            });

            this.conn.on('error', (err) => {
                cleanup();
                if (!this.connected) {
                    reject(new Error(`Error connecting to server: ${err.message}`));
                }
            });

            this.conn.on('close', () => {
                cleanup();
                this._handleDisconnect();
            });
        });
    }

    _handleDisconnect() {
        if (this.Quit) {
            this.connected = false;
            this.emit('disconnect');
            return;
        }

        this.connected = false;
        this.emit('disconnect');

        if (this.options.autoReconnect &&
            this.reconnectAttempts < this.options.maxReconnectAttempts) {

            this.reconnectAttempts++;
            const delay = this.options.reconnectDelay * Math.min(this.reconnectAttempts, 10);

            this.reconnectTimer = setTimeout(() => {
                this.emit('reconnecting', this.reconnectAttempts);
                this._attemptConnect().catch(() => {
                    // Connection errors are handled in _attemptConnect
                });
            }, delay);
        } else {
            this.emit('reconnect_failed');
        }
    }

    _flushPendingMessages() {
        while (this.pendingMessages.length > 0) {
            const { msg, resolve, reject } = this.pendingMessages.shift();
            this._sendInternal(msg).then(resolve).catch(reject);
        }
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
        this.options.autoReconnect = false;

        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        if (this.conn) {
            this.conn.end();
        }

        this.connected = false;
        this.emit('disconnect');
    }

    async send(msg) {
        if (!this.connected) {
            return new Promise((resolve, reject) => {
                if (this.options.autoReconnect) {
                    // Queue the message for when we reconnect
                    this.pendingMessages.push({ msg, resolve, reject });
                } else {
                    reject(new Error("Client not connected"));
                }
            });
        }

        return this._sendInternal(msg);
    }

    async _sendInternal(msg) {
        return new Promise((resolve, reject) => {
            try {
                const data = JSON.stringify(msg) + '\n';
                this.conn.write(data, (err) => {
                    if (err) {
                        this.emit('error', err);
                        reject(new Error(`Error sending message: ${err.message}`));
                    } else {
                        resolve();
                    }
                });
            } catch (err) {
                this.emit('error', err);
                reject(new Error(`Error encoding message: ${err.message}`));
            }
        });
    }

    async subscribe(topic, cb) {
        if (!this.connected && !this.options.autoReconnect) {
            throw new Error("Client not connected");
        }

        // Socket.IO style - allow multiple callbacks
        if (!this.subs_fun.has(topic)) {
            this.subs_fun.set(topic, []);

            // Only send SUB command if this is the first subscriber
            if (this.connected) {
                const msg = new Message({
                    cmd: 'SUB',
                    topic: topic
                });
                await this.send(msg);
            }
        }

        this.subs_fun.get(topic).push(cb);
    }

    async unsubscribe(topic, cb) {
        if (!this.subs_fun.has(topic)) {
            return;
        }

        if (cb) {
            // Remove specific callback
            const callbacks = this.subs_fun.get(topic);
            const index = callbacks.indexOf(cb);
            if (index !== -1) {
                callbacks.splice(index, 1);
            }

            if (callbacks.length === 0) {
                this.subs_fun.delete(topic);
                if (this.connected) {
                    const msg = new Message({
                        cmd: 'UNSUB',
                        topic: topic
                    });
                    await this.send(msg);
                }
            }
        } else {
            // Remove all callbacks
            this.subs_fun.delete(topic);
            if (this.connected) {
                const msg = new Message({
                    cmd: 'UNSUB',
                    topic: topic
                });
                await this.send(msg);
            }
        }
    }

    async publish(topic, payload) {
        let input = payload;
        if (typeof payload !== 'string') {
            try {
                input = JSON.stringify(payload);
            } catch (error) {
                this.emit('error', error);
                throw new Error(`Error serializing payload: ${error.message}`);
            }
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
        if (!this.connected && !this.options.autoReconnect) {
            throw new Error("Client not connected");
        }

        if (this.services.has(name)) {
            throw new Error(`Service '${name}' already registered`);
        }

        if (this.connected) {
            const msg = new Message({
                cmd: 'SER',
                topic: name
            });
            await this.send(msg);
        }

        this.services.set(name, cb);
    }

    async removeService(name) {
        if (!this.services.has(name)) {
            return;
        }

        this.services.delete(name);

        if (this.connected) {
            const msg = new Message({
                cmd: 'UNSER',
                topic: name
            });
            await this.send(msg);
        }
    }

    async request(name, payload, timeout = 5000) {
        if (!this.connected) {
            throw new Error("Client not connected");
        }

        let input = payload;
        if (typeof payload !== 'string') {
            try {
                input = JSON.stringify(payload);
            } catch (error) {
                this.emit('error', error);
                throw new Error(`Error serializing payload: ${error.message}`);
            }
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
                    try {
                        // Try to parse JSON if possible
                        const result = typeof resp.Payload === 'string' ?
                            JSON.parse(resp.Payload) : resp.Payload;
                        resolve(result);
                    } catch (e) {
                        resolve(resp.Payload);
                    }
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
                    this.emit('message', msg);

                    switch (msg.CMD) {
                        case 'PUB':
                            if (this.subs_fun.has(msg.Topic)) {
                                const funs = this.subs_fun.get(msg.Topic);
                                for (const funSub of funs) {
                                    let payload = msg.Payload;
                                    try {
                                        payload = JSON.parse(msg.Payload);
                                    } catch (error) {
                                        // Leave as string if not JSON
                                    }
                                    setImmediate(() => funSub(payload, msg.Topic_));
                                }
                            }
                            break;

                        case 'REQ':
                            if (this.services.has(msg.Topic)) {
                                const funService = this.services.get(msg.Topic);
                                setImmediate(() => {
                                    let payload = msg.Payload;
                                    try {
                                        payload = JSON.parse(msg.Payload);
                                    } catch (error) {
                                        // Leave as string if not JSON
                                    }

                                    funService(payload, (err, data) => {
                                        let output = data;
                                        try {
                                            output = JSON.stringify(data);
                                        } catch (error) {
                                            // Leave as is if can't stringify
                                        }

                                        this.send(new Message({
                                            cmd: 'RES',
                                            fromId: msg.FromID,
                                            payload: output,
                                            payload_err: err,
                                            topic: msg.Topic,
                                            reqId: msg.ReqID
                                        })).catch(err => this.emit('error', err));
                                    });
                                });
                            }
                            break;

                        case 'RES':
                            if (this.reqs.has(msg.ReqID)) {
                                const callback = this.reqs.get(msg.ReqID);
                                try {
                                    msg.Payload = typeof msg.Payload === 'string' ?
                                        JSON.parse(msg.Payload) : msg.Payload;
                                } catch (error) {
                                    // Leave as is if not JSON
                                }
                                callback(msg);
                            }
                            break;
                    }
                } catch (err) {
                    this.emit('error', new Error(`Error decoding message: ${err.message}`));
                }
            }
        });

        this.conn.on('error', (err) => {
            if (!this.Quit) {
                this.emit('error', err);
            }
        });

        this.conn.on('close', () => {
            this._handleDisconnect();
        });
    }
}

export { MQ, Message };