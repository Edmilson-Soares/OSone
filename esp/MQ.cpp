#include "MQ.h"
#include <Arduino.h>

MQ::MQ() : connected(false),
           reconnectInterval(5000),
           lastReconnectAttempt(0),
           port(1883)
{
}

void MQ::parseConnectionString(String connectionString)
{
    // Simple parser for mq://user:pass@host:port
    int schemeEnd = connectionString.indexOf("://");
    if (schemeEnd == -1 || connectionString.substring(0, schemeEnd) != "mq")
    {
        Serial.println("Invalid connection string format");
        return;
    }

    String rest = connectionString.substring(schemeEnd + 3);
    int atPos = rest.indexOf('@');

    if (atPos != -1)
    {
        String userPass = rest.substring(0, atPos);
        int colonPos = userPass.indexOf(':');
        if (colonPos != -1)
        {
            username = userPass.substring(0, colonPos);
            password = userPass.substring(colonPos + 1);
        }
        rest = rest.substring(atPos + 1);
    }

    int colonPos = rest.indexOf(':');
    if (colonPos != -1)
    {
        host = rest.substring(0, colonPos);
        port = rest.substring(colonPos + 1).toInt();
    }
    else
    {
        host = rest;
        port = 1883; // Default port
    }
}

bool MQ::sendMessage(JsonDocument &doc)
{
    if (!connected || !client.connected())
    {
        return false;
    }

    String output;
    serializeJson(doc, output);
    client.println(output);
    return true;
}

void MQ::processIncomingMessage(String message)
{
    DynamicJsonDocument doc(1024);
    DeserializationError error = deserializeJson(doc, message);

    if (error)
    {
        Serial.print("Failed to parse message: ");
        Serial.println(error.c_str());
        return;
    }

    String cmd = doc["cmd"].as<String>();
    String topic = doc["topic"].as<String>();
    String payload = doc["payload"].as<String>();
    String reqId = doc["reqId"].as<String>();
    String fromId = doc["fromId"].as<String>();
    String payload_err = doc["payload_err"].as<String>();
    String topic_ = doc["topic_"].as<String>();

    if (cmd == "PUB")
    {
        auto it = subscriptions.find(topic);
        if (it != subscriptions.end())
        {
            it->second(payload, topic_);
        }
    }
    else if (cmd == "REQ")
    {
        auto it = services.find(topic);
        if (it != services.end())
        {
            it->second(payload, [this, fromId, reqId, topic](String err, String data)
                       {
                DynamicJsonDocument response(256);
                response["cmd"] = "RES";
                response["fromId"] = fromId;
                response["payload"] = data;
                response["payload_err"] = err;
                response["topic"] = topic;
                response["reqId"] = reqId;
                
                sendMessage(response); });
        }
    }
    else if (cmd == "RES")
    {
        auto it = pendingRequests.find(reqId);
        if (it != pendingRequests.end())
        {
            it->second(payload, payload_err);
            pendingRequests.erase(it);
        }
    }
}

void MQ::maintainConnection()
{
    if (!client.connected())
    {
        if (millis() - lastReconnectAttempt > reconnectInterval)
        {
            lastReconnectAttempt = millis();
            if (connect())
            {
                Serial.println("Reconnected to MQ server");
            }
        }
    }
}

bool MQ::connect(String connectionString)
{
    url = connectionString;
    parseConnectionString(url);
    return connect();
}

bool MQ::connect()
{
    if (client.connect(host.c_str(), port))
    {
        connected = true;

        // Authenticate
        DynamicJsonDocument authMsg(256);
        authMsg["cmd"] = "AUTH";
        authMsg["topic"] = username;
        authMsg["payload"] = password;
        authMsg["reqId"] = String(millis());

        if (!sendMessage(authMsg))
        {
            connected = false;
            return false;
        }

        // TODO: Wait for auth response and set clientId
        return true;
    }
    return false;
}

void MQ::disconnect()
{
    client.stop();
    connected = false;
}

void MQ::loop()
{
    maintainConnection();

    while (client.available())
    {
        String line = client.readStringUntil('\n');
        processIncomingMessage(line);
    }

    // Process queued messages
    while (!messageQueue.empty() && connected)
    {
        String msg = messageQueue.front();
        client.println(msg);
        messageQueue.pop();
    }
}

bool MQ::subscribe(String topic, std::function<void(String, String)> callback)
{
    if (!connected)
        return false;

    DynamicJsonDocument subMsg(128);
    subMsg["cmd"] = "SUB";
    subMsg["topic"] = topic;

    if (sendMessage(subMsg))
    {
        subscriptions[topic] = callback;
        return true;
    }
    return false;
}

bool MQ::publish(String topic, String payload)
{
    DynamicJsonDocument pubMsg(256);
    pubMsg["cmd"] = "PUB";
    pubMsg["topic"] = topic;
    pubMsg["payload"] = payload;
    pubMsg["reqId"] = String(millis());

    if (connected)
    {
        return sendMessage(pubMsg);
    }
    else
    {
        String output;
        serializeJson(pubMsg, output);
        messageQueue.push(output);
        return true;
    }
}

bool MQ::registerService(String name, std::function<void(String, std::function<void(String, String)>)> callback)
{
    if (!connected)
        return false;

    DynamicJsonDocument serMsg(128);
    serMsg["cmd"] = "SER";
    serMsg["topic"] = name;

    if (sendMessage(serMsg))
    {
        services[name] = callback;
        return true;
    }
    return false;
}

bool MQ::request(String name, String payload, unsigned long timeout,
                 std::function<void(String, String)> callback)
{
    if (!connected)
        return false;

    String reqId = String(millis());
    DynamicJsonDocument reqMsg(256);
    reqMsg["cmd"] = "REQ";
    reqMsg["topic"] = name;
    reqMsg["payload"] = payload;
    reqMsg["reqId"] = reqId;

    if (sendMessage(reqMsg))
    {
        pendingRequests[reqId] = callback;

        unsigned long startTime = millis();
        while (millis() - startTime < timeout)
        {
            loop();
            if (pendingRequests.find(reqId) == pendingRequests.end())
            {
                return true; // Response received
            }
            delay(10);
        }

        // Timeout reached
        pendingRequests.erase(reqId);
        callback("", "timeout");
        return false;
    }
    return false;
}