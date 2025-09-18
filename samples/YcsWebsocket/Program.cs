using System.Net;
using YcsWebsocket;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

var app = builder.Build();

// ���WebSocket֧��
app.UseWebSockets();
app.UseYjsWebSocketServer();

app.UseHttpsRedirection();


// ����HTTP GET����
app.MapGet("/", async context =>
{
    context.Response.StatusCode = (int)HttpStatusCode.OK;
    context.Response.ContentType = "text/plain";
    await context.Response.WriteAsync("okay");
});

// ����WebSocket����
app.Use(async (context, next) =>
{
    if (context.WebSockets.IsWebSocketRequest)
    {
        // ������������֤�߼�
        // ������cookies��URL����
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
