using System.Net;
using YcsWebsocket;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

var app = builder.Build();

// 添加WebSocket支持
app.UseWebSockets();
app.UseYjsWebSocketServer();

app.UseHttpsRedirection();


// 处理HTTP GET请求
app.MapGet("/", async context =>
{
    context.Response.StatusCode = (int)HttpStatusCode.OK;
    context.Response.ContentType = "text/plain";
    await context.Response.WriteAsync("okay");
});

// 处理WebSocket连接
app.Use(async (context, next) =>
{
    if (context.WebSockets.IsWebSocketRequest)
    {
        // 这里可以添加认证逻辑
        // 例如检查cookies或URL参数
        /*
        if (!Authenticate(context))
        {
            context.Response.StatusCode = (int)HttpStatusCode.Unauthorized;
            return;
        }
        */

        var webSocket = await context.WebSockets.AcceptWebSocketAsync();
        await YjsWebSocketServer.SetupWSConnection(webSocket, context);
    }
    else
    {
        await next();
    }
});
app.Run();
