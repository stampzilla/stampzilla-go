function WebSocketTest()
{
  if ("WebSocket" in window)
  {
     // Let us open a web socket
     var ws = new WebSocket("ws://"+location.host+"/socket");
     ws.onopen = function()
     {
        // Web Socket is connected, send data using send()
        ws.send("Message to send");
        console.log("Message is sent...");
     };
     ws.onmessage = function (evt) 
     { 
        var received_msg = evt.data;
        //console.log("Message is received...",received_msg);

        obj = JSON.parse(received_msg);
        eval(obj.Msg);
     };
     ws.onclose = function()
     { 
        // websocket is closed.
        console.log("Connection is closed..."); 
     };
  }
  else
  {
     // The browser doesn't support WebSocket
     console.log("WebSocket NOT supported by your Browser!");
  }
}
