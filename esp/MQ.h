#ifndef MQ_CLIENT_H
#define MQ_CLIENT_H

#include <WiFiClient.h>
#include <ArduinoJson.h>
#include <map>
#include <functional>
#include <vector>
#include <queue>

class MQ
{
private:
    WiFiClient client;
    bool connected;
    String url;
    String clientId;

    std::map<String, std::function<void(String, String)>> subscriptions;
    std::map<String, std::function<void(String, std::function<void(String, String)>)>> services;
    std::map<String, std::function<void(String, String)>> pendingRequests;

    std::queue<String> messageQueue;
    unsigned long lastReconnectAttempt;
    const unsigned long reconnectInterval;

    String username;
    String password;
    String host;
    int port;

    void parseConnectionString(String connectionString);
    bool sendMessage(JsonDocument &doc);
    void processIncomingMessage(String message);
    void maintainConnection();

public:
    MQ();
    bool connect(String connectionString);
    bool connect();
    void disconnect();
    void loop();
    bool subscribe(String topic, std::function<void(String, String)> callback);
    bool publish(String topic, String payload);
    bool registerService(String name, std::function<void(String, std::function<void(String, String)>)> callback);
    bool request(String name, String payload, unsigned long timeout, std::function<void(String, String)> callback);
};

#endif