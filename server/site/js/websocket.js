(function () {
    "use strict";
//function WebSocketTest()
//{
  //if ("WebSocket" in window)
  //{
     //// Let us open a web socket
     //var ws = new WebSocket("ws://"+location.host+"/socket");
     //ws.onopen = function()
     //{
        //// Web Socket is connected, send data using send()
        //ws.send("Message to send");
        //console.log("Message is sent...");
     //};
     //ws.onmessage = function (evt) 
     //{ 
        //var received_msg = evt.data;
        ////console.log("Message is received...",received_msg);

        //obj = JSON.parse(received_msg);
        //eval(obj.Msg);
     //};
     //ws.onclose = function()
     //{ 
        //// websocket is closed.
        //console.log("Connection is closed..."); 
     //};
  //}
  //else
  //{
     //// The browser doesn't support WebSocket
     //console.log("WebSocket NOT supported by your Browser!");
  //}
//}



    if ("WebSocket" in window){
        var ws = function () {
            _.extend(this, Backbone.Events);
            var self = this;

            self.socket = new WebSocket("ws://" + location.host + "/socket");
            console.log("Using a standard websocket");

            self.socket.onopen = function(e) {
                self.trigger('open', e);
                console.log('socket opened');
            };
            self.socket.onerror = function(e) {
                self.trigger('error', e);
            };
            self.socket.onmessage = function(e) {
                //self.trigger('message', e);
                //self.trigger('data', e.data);
                //console.log(e.data);
                self.trigger('change_ws', JSON.parse(e.data));
            };
            self.socket.onclose = function(e) {
                self.trigger('close', e);
                console.log('socket closed');
            };
        };  
        Stampzilla.Websocket = new ws();
    }


}());
