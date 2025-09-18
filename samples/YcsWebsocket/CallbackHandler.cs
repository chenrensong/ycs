namespace YcsWebsocket
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Text;
    using System.Text.Json;
    using System.Threading;
    using System.Threading.Tasks;

    public class CallbackHandler
    {
        private static readonly Uri CallbackUrl = Environment.GetEnvironmentVariable("CALLBACK_URL") != null
            ? new Uri(Environment.GetEnvironmentVariable("CALLBACK_URL"))
            : null;

        private static readonly int CallbackTimeout = int.TryParse(
            Environment.GetEnvironmentVariable("CALLBACK_TIMEOUT") ?? "5000",
            out var timeout) ? timeout : 5000;

        private static readonly Dictionary<string, string> CallbackObjects =
            Environment.GetEnvironmentVariable("CALLBACK_OBJECTS") != null
                ? JsonSerializer.Deserialize<Dictionary<string, string>>(
                    Environment.GetEnvironmentVariable("CALLBACK_OBJECTS"))
                : new Dictionary<string, string>();

        public static bool IsCallbackSet => CallbackUrl != null;

        private static readonly HttpClient HttpClient = new HttpClient
        {
            Timeout = TimeSpan.FromMilliseconds(CallbackTimeout)
        };

        public static async Task CallbackHandlerAsync(WSSharedDoc doc)
        {
            if (CallbackUrl == null) return;

            var room = doc.Name;
            var dataToSend = new
            {
                room,
                data = new Dictionary<string, object>()
            };

            foreach (var (sharedObjectName, sharedObjectType) in CallbackObjects)
            {
                dataToSend.data[sharedObjectName] = new
                {
                    type = sharedObjectType,
                    content = GetContent(sharedObjectName, sharedObjectType, doc)
                };
            }

            await CallbackRequestAsync(CallbackUrl, dataToSend);
        }

        private static async Task CallbackRequestAsync(Uri url, object data)
        {
            try
            {
                var jsonData = JsonSerializer.Serialize(data);
                var content = new StringContent(jsonData, Encoding.UTF8, "application/json");

                var response = await HttpClient.PostAsync(url, content);

                if (!response.IsSuccessStatusCode)
                {
                    Console.Error.WriteLine($"Callback request failed with status code: {response.StatusCode}");
                }
            }
            catch (TaskCanceledException)
            {
                Console.WriteLine("Callback request timed out.");
            }
            catch (Exception e)
            {
                Console.Error.WriteLine($"Callback request error: {e.Message}");
            }
        }

        private static object GetContent(string objName, string objType, WSSharedDoc doc)
        {
            return objType switch
            {
                "Array" => doc.GetArray(objName),
                "Map" => doc.GetMap(objName),
                "Text" => doc.GetText(objName),
                "XmlFragment" => doc.GetXmlFragment(objName),
                "XmlElement" => doc.GetXmlElement(objName),
                _ => new object()
            };
        }
    }

}
