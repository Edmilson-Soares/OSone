from  mq.client import MQ, Callable
import time
# Create client
client = MQ()



# Example usage
def message_handler(msg: str, topic: str):
    print(f"Received on {topic}: {msg}")

def service_handler(data: str, reply: Callable[[str, str], None]):
    print(f"Service received: {data}")
    reply("", f"Processed: {data}")
#37.27.39.202:
try:
    # Connect to server
    client.connect("mq://root:chave-secreta-32-bytes-123456789@127.0.0.1:4052")
    
    # Subscribe to topic
    client.subscribe("test.topic", message_handler)
    
    # Publish message
    client.publish("test.topic", "Hello World!")
    
    # Register service
    client.service("echo", service_handler)
    
    # Make request
    response, err = client.request("echo", "test data", 5.0)
    if err:
        print(f"Error: {err}")
    else:
        print(f"Response: {response}")
        
    # Keep running
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    pass
finally:
    client.disconnect()