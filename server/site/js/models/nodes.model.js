/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
(function () {
    "use strict";

    Stampzilla.Models.NodeModel = Backbone.Model.extend({
        urlRoot : '/api/node',
        idAttribute: "Id",
        //initialize:function () {
        //}
    });

    Stampzilla.Collections.NodesCollection = Backbone.Collection.extend({
        model: Stampzilla.Models.NodeModel,
        url : '/api/nodes',
        initialize : function() {
            this.deferred = this.fetch();
        }
    });

}());
