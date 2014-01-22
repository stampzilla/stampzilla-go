/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
(function () {
    "use strict";

    Stampzilla.Collections.StatesCollection = Backbone.Collection.extend({
        model: Stampzilla.Models.StateModel,
    });

    Stampzilla.Models.StateModel = Backbone.Model.extend({
        url: function(){
            //return '/api/'+this.node.get('Id')+'/state';
            return '/api/node/'+this.node.get('Id')+'/state';
        },
        initialize : function(attributes,options) {
            this.node = options.node;
        }
    });

    Stampzilla.Models.NodeModel = Backbone.Model.extend({
        urlRoot : '/api/node',
        idAttribute: "Id",
        initialize: function(){
            var self = this;

            //Update our model when we get change from websocket
            Stampzilla.Websocket.on("change_ws", function(data){
                self.set(self.parse(data));
            });

        },
        parse: function (data){

            if(data.State instanceof Object){
                data.State = new Stampzilla.Models.StateModel(_.clone(data.State),{node:this});
            }
            return data;
        }
    });

    Stampzilla.Collections.NodesCollection = Backbone.Collection.extend({
        model: Stampzilla.Models.NodeModel,
        url : '/api/nodes',
        initialize : function() {
            this.deferred = this.fetch();
        }
    });





}());
