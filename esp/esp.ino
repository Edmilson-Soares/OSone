#include <WiFi.h>
#include "MQ.h"

const char *ssid = "your_SSID";
const char *password = "your_WiFi_password";
String mqUrl = "mq://user:pass@server:port";

MQ mqClient;

void setup()
{
    Serial.begin(115200);

    // Connect to WiFi
    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED)
    {
        delay(500);
        Serial.print(".");
    }
    Serial.println("WiFi connected");

    // Connect to MQ server
    if (mqClient.connect(mqUrl))
    {
        Serial.println("Connected to MQ server");

        // Subscribe to a topic
        mqClient.subscribe("test/topic", [](String msg, String topic)
                           {
            Serial.print("Received message on ");
            Serial.print(topic);
            Serial.print(": ");
            Serial.println(msg); });

        // Register a service
        mqClient.registerService("echo", [](String data, std::function<void(String, String)> reply)
                                 {
                                     reply("", data); // Echo back the data
                                 });
    }
    else
    {
        Serial.println("Failed to connect to MQ server");
    }
}

void loop()
{
    mqClient.loop();

    // Example: Publish a message every 5 seconds
    static unsigned long lastPublish = 0;
    if (millis() - lastPublish > 5000)
    {
        lastPublish = millis();
        mqClient.publish("test/topic", "Hello from ESP " + String(millis()));

        // Example request
        mqClient.request("echo", "ping", 1000, [](String err, String data)
                         {
            if (err != "") {
                Serial.print("Error: ");
                Serial.println(err);
            } else {
                Serial.print("Response: ");
                Serial.println(data);
            } });
    }
}