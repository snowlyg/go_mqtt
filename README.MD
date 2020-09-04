### 配置文件

```json
{"topic":"replace","qos":0,"clientid":"replace","broker":"replace","user_name":"replace","password":"replace","num":1}
```
```json
{"topic":"iot/device/123/control","qos":2,"clientid":"7d7c43db-d061-42cf-9e9f-5c9e36b1dd34","broker":"tcp://10.0.0.27:1883","user_name":"Chindeo","password":"P@ssw0rd","num":50}
```

```yaml
topic: 'iot/device/123/control' 
qos: 2 
clientid: '7d7c43db-d061-42cf-9e9f-5c9e36b1dd34' 
broker: 'tcp://10.0.0.27:1883' 
user_name: 'Chindeo' 
password: 'P@ssw0rd' 
num: 50 
```

### udp 测试 端口 10006

```shell script
cd udp & go run main.go
#输入 json 修改配置
```