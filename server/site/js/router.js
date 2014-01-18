/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
window.Stampzilla = window.Stampzilla || {Routers: {}, Collections: {}, Models: {}, Views: {}, Data: {}, Websocket: {}};
(function () {
    "use strict";
    Stampzilla.loadTemplates = function(views,callback){
        var deferreds = [];
        $.each(views, function(index, view) {
            if (Stampzilla.Views[view]) {
                deferreds.push($.get('templates/' + view + '.html', function(data,textStatus) {
                    console.log(textStatus);
                    Stampzilla.Views[view].prototype.template = _.template(data);
                }, 'html').fail(function(){
                    console.log("failed to fetch " + this.url);
                }));
            } else {
                alert(view + " not found");
            }
        });

        $.when.apply(null, deferreds).done(callback).fail(callback);
    }

    Stampzilla.registerModel = function(model){


    }
    Stampzilla.registerView = function(viewName, viewObject){
        
        if(Stampzilla.Views[viewName] === undefined){
            Stampzilla.Views[viewName] =  viewObject;
        }
        return Stampzilla.Views[viewName];
    }
    Stampzilla.getViews = function(){
        var ret = new Array();
        _.each(Stampzilla.Views, function(view,key){
            ret.push(key);
        });
        return ret;
    }
    Stampzilla.getView = function(viewName){
        return Stampzilla.Views[viewName] || {};
    }

    Stampzilla.Routers.MainRouter = Backbone.Router.extend({
        routes: {
            "": "index",
            "nodes/": "nodes",
            "node/:id": "showNode",
        },

        initialize: function () {
            Stampzilla.Data.NodesCollection = new Stampzilla.Collections.NodesCollection();
            this.cached = {
                view:{},
                model:{}
            }
        },

        nodes: function () {
            this.cached.view.NodesTable = this.cached.view.NodesTable || new Stampzilla.Views.NodesTable({collection:Stampzilla.Data.NodesCollection});
            
            $('#main').html(this.cached.view.NodesTable.el);
            this.cached.view.NodesTable.render();
        },
        showNode: function (id) {
            //wait for intial fetch to finish using deferred before we get model by id
            var self = this;
            Stampzilla.Data.NodesCollection.deferred.done(function(){
                self.cached.view.Node = self.cached.view.Node || new Stampzilla.Views.Node({model:Stampzilla.Data.NodesCollection.get(id)});
                $('#main').html(self.cached.view.Node.el);
                self.cached.view.Node.render();
            });
        },
        index: function () {
            //this.currentView = new Stampzilla.Views.NodesView();
            $('#main').html("");;
        }
    });



}());
