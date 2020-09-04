### 配置文件

```json
{"qos":2,"clientid":"7d7c43db-d061-42cf-9e9f-5c9e36b1dd34","broker":"tcp://10.0.0.27:1883","user_name":"Chindeo","password":"P@ssw0rd","num":50}
```

```yaml
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

## msg
```json
{"type":"cmd","data":{"text":"go","arg":"env"}}
```

### 编译参数
```shell script
GOOS=linux;GOARCH=arm;GOARM=7;CC=arm-linux-gnueabi-gcc-4.7
```