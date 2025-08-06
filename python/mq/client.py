import socket
import threading
import queue
import time
import json
from urllib.parse import urlparse
from typing import Dict, List, Callable, Optional, Tuple, NamedTuple, Any

class MQError(Exception):
    """Base exception for all MQ client errors"""
    pass

class MQConnectionError(MQError):
    """Exception raised for connection related errors"""
    pass

class MQTimeoutError(MQError):
    """Exception raised when an operation times out"""
    pass

class MQProtocolError(MQError):
    """Exception raised for protocol violations"""
    pass

class ConnectionInfo(NamedTuple):
    """Parsed connection information"""
    username: str
    password: str
    host: str
    port: str

class Message:
    """MQ message structure"""
    
    def __init__(
        self,
        cmd: str,
        topic: str,
        payload: str = "",
        req_id: str = "",
        payload_err: str = "",
        from_id: str = "",
        topic_: str = ""
    ):
        self.cmd = cmd
        self.topic = topic
        self.payload = payload
        self.req_id = req_id
        self.payload_err = payload_err
        self.from_id = from_id
        self.topic_ = topic_

    def to_dict(self) -> Dict:
        """Convert message to dictionary"""
        data = {
            "cmd": self.cmd,
            "topic": self.topic,
            "payload": self.payload,
        }
        
        # Optional fields
        if self.req_id:
            data["reqId"] = self.req_id
        if self.payload_err:
            data["payload_err"] = self.payload_err
        if self.from_id:
            data["fromId"] = self.from_id
        if self.topic_:
            data["topic_"] = self.topic_
            
        return data

    def to_json(self) -> str:
        """Serialize message to JSON"""
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> 'Message':
        """Create Message from JSON string"""
        try:
            data = json.loads(json_str)
            return cls(
                cmd=data.get("cmd", ""),
                topic=data.get("topic", ""),
                payload=data.get("payload", ""),
                req_id=data.get("reqId", ""),
                payload_err=data.get("payload_err", ""),
                from_id=data.get("fromId", ""),
                topic_=data.get("topic_", "")
            )
        except json.JSONDecodeError as e:
            raise MQProtocolError(f"Invalid message format: {e}")

class MQ:
    """MQ client implementation"""
    
    def __init__(self):
        self.conn: Optional[socket.socket] = None
        self.reqs: Dict[str, queue.Queue] = {}
        self.lock = threading.RLock()
        self.url = ""
        self.id = ""
        self.connected = False
        self.subs_fun: Dict[str, List[Callable[[str, str], None]]] = {}
        self.services: Dict[str, Callable[[str, Callable[[str, str], None]], None]] = {}
        self.quit_event = threading.Event()
        self.reader_thread: Optional[threading.Thread] = None

    def _parse_connection(self, connection_string: str) -> ConnectionInfo:
        """Parse connection URL in format mq://username:password@host:port"""
        try:
            u = urlparse(connection_string)
            if u.scheme != "mq":
                raise ValueError("Invalid scheme, expected 'mq'")

            username = u.username or ""
            password = u.password or ""
            host = u.hostname or "localhost"
            port = u.port or "4222"  # Default MQ port

            if not host:
                raise ValueError("Missing host in connection string")

            return ConnectionInfo(username, password, host, str(port))
        except Exception as e:
            raise MQConnectionError(f"Failed to parse connection string: {e}")

    def connect(self, url: str) -> None:
        """Connect to MQ server"""
        self.url = url
        try:
            auth = self._parse_connection(self.url)
            self.conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.conn.connect((auth.host, int(auth.port)))
            self.connected = True
        except Exception as e:
            raise MQConnectionError(f"Error connecting to server: {e}")

        # Start reader thread
        self.reader_thread = threading.Thread(
            target=self._read_loop,
            daemon=True,
            name="MQ-Reader"
        )
        self.reader_thread.start()

        # Authenticate
        self._auth(auth.username, auth.password, 5)

    def _auth(self, username: str, password: str, timeout: float) -> None:
        """Authenticate with the server"""
        req_id = str(int(time.time() * 1e9))
        msg = Message(
            cmd="AUTH",
            topic=username,
            payload=password,
            req_id=req_id
        )

        self._send(msg)

        # Create response queue
        resp_queue = queue.Queue(maxsize=1)

        with self.lock:
            self.reqs[req_id] = resp_queue

        try:
            # Wait for response with timeout
            try:
                resp = resp_queue.get(timeout=timeout)
                if resp.payload_err:
                    raise MQConnectionError(resp.payload_err)
                self.id = resp.payload
            except queue.Empty:
                raise MQTimeoutError(f"Timeout after {timeout} seconds")
        finally:
            with self.lock:
                if req_id in self.reqs:
                    del self.reqs[req_id]

    def disconnect(self) -> None:
        """Disconnect from server"""
        self.quit_event.set()
        if self.conn:
            try:
                self.conn.close()
            except:
                pass
        self.connected = False
        if self.reader_thread:
            self.reader_thread.join(timeout=1)

    def _send(self, msg: Message) -> None:
        """Send a message to the server"""
        if not self.conn:
            raise MQConnectionError("Not connected to server")

        try:
            data = msg.to_json() + "\n"
            self.conn.sendall(data.encode('utf-8'))
        except Exception as e:
            raise MQConnectionError(f"Error sending message: {e}")

    def subscribe(self, topic: str, callback: Callable[[str, str], None]) -> None:
        """Subscribe to a topic"""
        if not self.connected:
            raise MQConnectionError("Client not connected")

        msg = Message(cmd="SUB", topic=topic)
        self._send(msg)

        with self.lock:
            if topic not in self.subs_fun:
                self.subs_fun[topic] = []
            self.subs_fun[topic].append(callback)

    def publish(self, topic: str, payload: str) -> None:
        """Publish a message to a topic"""
        if not self.connected:
            raise MQConnectionError("Client not connected")

        msg = Message(
            cmd="PUB",
            topic=topic,
            payload=payload,
            req_id=str(int(time.time() * 1e9))
        )
        self._send(msg)

    def service(self, name: str, callback: Callable[[str, Callable[[str, str], None]], None]) -> None:
        """Register a service"""
        if not self.connected:
            raise MQConnectionError("Client not connected")

        msg = Message(cmd="SER", topic=name)
        self._send(msg)

        with self.lock:
            self.services[name] = callback

    def request(self, name: str, payload: str, timeout: float) -> Tuple[str, Optional[str]]:
        """Make a request and wait for response"""
        if not self.connected:
            raise MQConnectionError("Client not connected")

        req_id = str(int(time.time() * 1e9))
        msg = Message(
            cmd="REQ",
            topic=name,
            payload=payload,
            req_id=req_id
        )

        self._send(msg)

        # Create response queue
        resp_queue = queue.Queue(maxsize=1)

        with self.lock:
            self.reqs[req_id] = resp_queue

        try:
            # Wait for response with timeout
            try:
                resp = resp_queue.get(timeout=timeout)
                if resp.payload_err:
                    return "", resp.payload_err
                return resp.payload, None
            except queue.Empty:
                return "", f"Timeout after {timeout} seconds"
        finally:
            with self.lock:
                if req_id in self.reqs:
                    del self.reqs[req_id]

    def _read_loop(self) -> None:
        """Read messages from server in a loop"""
        if not self.conn:
            return

        buffer = ""
        while not self.quit_event.is_set():
            try:
                data = self.conn.recv(4096)
                if not data:  # Connection closed
                    self.connected = False
                    break

                buffer += data.decode('utf-8')

                while "\n" in buffer:
                    line, buffer = buffer.split("\n", 1)
                    if not line:
                        continue

                    try:
                        msg = Message.from_json(line)
                        self._handle_message(msg)
                    except MQProtocolError as e:
                        print(f"Protocol error: {e}")
                    except Exception as e:
                        print(f"Error processing message: {e}")

            except ConnectionError:
                self.connected = False
                break
            except Exception as e:
                print(f"Error in read loop: {e}")
                self.connected = False
                break

    def _handle_message(self, msg: Message) -> None:
        """Handle incoming message"""
        with self.lock:
            if msg.cmd == "PUB":
                if msg.topic in self.subs_fun:
                    for callback in self.subs_fun[msg.topic]:
                        try:
                            callback(msg.payload, msg.topic_)
                        except Exception as e:
                            print(f"Error in subscription callback: {e}")

            elif msg.cmd == "REQ":

                if msg.topic in self.services:
                    def reply(err: str, data: str) -> None:
                        reply_msg = Message(
                            cmd="RES",
                            from_id=msg.from_id,
                            payload=data,
                            payload_err=err,
                            topic=msg.topic,
                            req_id=msg.req_id
                        )
                        self._send(reply_msg)

                    try:
                        self.services[msg.topic](msg.payload, reply)
                    except Exception as e:
                        print(f"Error in service callback: {e}")
                        reply(str(e), "")

            elif msg.cmd == "RES":
                if msg.req_id in self.reqs:
                    try:
                        self.reqs[msg.req_id].put(msg)
                    except queue.Full:
                        pass

    def unsubscribe(self, topic: str) -> None:
        """Unsubscribe from a topic"""
        with self.lock:
            if topic in self.subs_fun:
                del self.subs_fun[topic]


