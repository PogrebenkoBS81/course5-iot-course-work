## To build images:
```
docker build . -t device-event-multiplier

docker build . -t receiver-mock
```

## To run service type:
```docker-compose up```

## Test input
### 1) http to http handler
```curl -X POST -H "Content-Type: application/json" -d '{ "DeviceId": "123", "Time": 2, "Temperature": 30.30, "Humidity": 20.20}' http://localhost:8080/start/http```

### 2) http to pub sub handler
```curl -X POST -H "Content-Type: application/json" -d '{ "DeviceId": "123", "Time": 1, "Temperature": 30.30, "Humidity": 20.20}' http://localhost:8080/start/kafka```

### 3) pub sub to pub sub handler
```docker exec -it <kafka-container-name> /bin/bash```

```echo '{ "DeviceId": "123", "Time": 600, "Temperature": 30.30, "Humidity": 20.20}' | kafka-console-producer --broker-list localhost:9092 --topic device_data_test```