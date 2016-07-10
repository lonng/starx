# 第一个StarX应用，HelloWorld

## 服务器

1. 在`$GOPATH/src`下面新建一个文件夹`hello_world`, 复制StarX目录下的`config-examples`到新建的目录，并把文件夹名字改为`configs`
2. 新建`main.go`
```
package main

import (
	"github.com/chrislonng/starx"
)

type HelloWorld struct {}

// 下面四个方法是实现Component接口，是框架对Component的生命周期控制
func (h *HelloWorld) Init(){}
func (h *HelloWorld) AfterInit(){}
func (h *HelloWorld) BeforeShutdown(){}
func (h *HelloWorld) Shutdown(){}

// 以下方法是客户端可以请求的接口，方法的签名第一个参数是请求这个服务的Session
// 第二个参数是请求的数据，请求的数据需要自己反序列化，因为不同的开发者可能使
// 用json、msgpack、protobuf。
func (h *HelloWorld) Hello(session *starx.Session, data []byte) error {
    session.Response([]byte("World")
    return nil
}

// 需要新增客户端可以请求的接口，只需要在相应的结构上新增公开的方法，不需要做任何
// 额外的工作，客户端就可以请求
func (h *HelloWorld) World(session *starx.Session, data []byte) error {
    session.Response([]byte("World")
    session.Push("OnChatRoomClose", []byte(`{"roomID": 222222}`))
    return nil
}

func main() {
	starx.Set("demo", func(){
		starx.Handler(new(HelloWorld))
	});
	starx.Start()
}
```
到目前为止，我们的hello_world目录应该是这样的：
```
.
├── configs
│   ├── app.json
│   └── servers.json
└── main.go
```
3. 打开控制台，在`hello_world`目录下面`go build`
4. ./hello_world demo-server-1

恭喜你，你的第一个StarX应用已经成功跑起来了，具体代码的解释留在最后，我们先看看客户端怎么写。

## 客户端

目前客户端有一个C#的SDK，包含在[starx-demo-unity](https://github.com/chrislonng/starx-demo-unity)，客户端测试代码可以是一个Unity3D工程，也可以是一个VS的C#工程，这里我们只展示主要的代码，具体代码可以参考`starx-demo-unity`

```
StarXClient client = new StarXClient();
client.Init("127.0.0.1", 3250, () =>
{
    Debug.Log("init client callback");
    client.Connect((data) =>
    {
        Debug.Log("connect client callback");

        // 请求demo服务器的HelloWorld.Hello接口
        client.Request("demo.HelloWorld.Hello", Encoding.UTF8.GetBytes("hello world test"), (resp) =>
        {
            Debug.Log("demo.HelloWorld.Hello: " + Encoding.UTF8.GetString(resp));
        });

        // 请求demo服务器的HelloWorld.Hello接口
        // 如果请求的是当前服务器，可以省略服务器类型名，当服务器为集群的时候，请求非前端服务器，需要添加服务器类型名
        client.Request("HelloWorld.Hello", Encoding.UTF8.GetBytes("hello world test"), (resp) =>
        {
            Debug.Log("HelloWorld.Hello: " + Encoding.UTF8.GetString(resp));
        });

	// 服务器主动推送消息
	client.On("OnChatRoomClose", (m) =>
	{
		Debug.Log("OnChatRoomClose: " + Encoding.UTF8.GetString(m));
	});

        // 通知gate服务器，LoginHandler.NotifyTest服务
	client.Notify("HelloWorld.World", Encoding.UTF8.GetBytes("notify test message"));
    });
});
```

如果没有意外，你会看到`demo.HelloWorld.Hello: World`的日志

## 解释
前后端通信基本完成，接下来我们对上面的代码作一个简略的解释

### 配置
配置使用json格式，包含两个配置`app.json`和`servers.json`

1. app.json包含应用运行的配置，这里目前只有一个应用名和Standalone，如果Standalone为true，则应用以单节点的模式运行
2. servers.json包含所有服务器的信息，新增服务器只需要修改这个配置，重启即可
```
{
  "gate": [
    {"id": "gate-server-1","host": "0.0.0.0", "port": 3250, "isFrontend": true},
    {"id": "gate-server-2","host": "0.0.0.0", "port": 3251, "isFrontend": true},
    {"id": "gate-server-3","host": "0.0.0.0", "port": 3252, "isFrontend": true},
    {"id": "gate-server-4","host": "0.0.0.0", "port": 3253, "isFrontend": true}
  ],
  "chat": [
    {"id": "chat-server-1","host": "127.0.0.1", "port": 3260, "isFrontend": false},
    {"id": "chat-server-2","host": "127.0.0.1", "port": 3261, "isFrontend": false},
    {"id": "chat-server-3","host": "127.0.0.1", "port": 3262, "isFrontend": false},
    {"id": "chat-server-4","host": "127.0.0.1", "port": 3263, "isFrontend": false},
  ]
}
```
上面是一个`servers.json`的示例，`gate``chat`是服务器的类型，然后对应每个类型的服务器配置，其中id是这台服务器的唯一标识，如果重复，后面会覆盖前面的配置，host、port没什么好解释的，isFrontend这个参数表示这台服务器是否是前端服务器，前端服务器才能接收客户端的连接，后端服务器一般都是内网服务器，不对外提供服务。

回到我们的hello_world示例，我们这里只有一个demo类型的服务器，id为demo-server-1，监听0.0.0.0:3250，是前端服务器，我们使用命令`./hello_world demo-server-1`启动服务器，其中demo-server-1就是这个服务器进程的配置

### 服务器代码

1. 在服务器，我们新建了一个HelloWorld结构，并实现了Component接口，并新增了Hello和World方法，所以我们在客户端就可以通过`"demo.HelloWorld.Hello"`和`"demo.HelloWorld.World"`来请求，所有新增在HelloWorld上的公开方法，客户端都可以请求。
2. starx.Set("serverType", callback)这里是不同服务器的初始化设置，第一个参数为服务器类型，可以针对不同服务器，注册不同的服务
3. 使用starx.Handler(new(HelloWorld))来注册HelloWorld这个服务
4. starx.Start()来运行StarX应用

客户端 ---> 服务器
1. request 需要一个回调函数，在服务器对这个消息response的时候调用
2. notify 不需要回调函数
服务器 ---> 客户端
1. response 响应一个客户端请求
2. push 主动推送数据到客户端，session.Push("OnChatRoomClose", []byte(`{"roomID": 222222}`))

客户端这样注册push的回调函数
```
client.On("OnChatRoomClose", callback)
```

### 客户端代码
客户端代码很容易理解



