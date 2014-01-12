/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
window.Stampzilla = window.Stampzilla || {Routers: {}, Collections: {}, Models: {}, Views: {}, Data: {}};
(function () {
    "use strict";
    Stampzilla.loadTemplates = function(views,callback){
        var deferreds = [];
        $.each(views, function(index, view) {
            if (Stampzilla.Views[view]) {
                deferreds.push($.get('templates/' + view + '.html', function(data) {
                    Stampzilla.Views[view].prototype.template = _.template(data);
                }, 'html'));
            } else {
                alert(view + " not found");
            }
        });

        $.when.apply(null, deferreds).done(callback);
    }
    Stampzilla.Routers.MainRouter = Backbone.Router.extend({
        routes: {
            "": "index",
            "nodes/": "nodes",
            "node/:id": "showNode",
        },

        initialize: function () {
            Stampzilla.Data.NodesCollection = new Stampzilla.Collections.NodesCollection();
            //Stampzilla.Data.NodesCollection.fetch();

            this.cached = {
                view:{},
                model:{}
            }
        },

        index: function () {
            this.cached.view.NodesTable = this.cached.view.NodesTable || new Stampzilla.Views.NodesTable({collection:Stampzilla.Data.NodesCollection});
            //Stampzilla.Data.NodesCollection.fetch();
            
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
        nodes: function () {
            //this.currentView = new Stampzilla.Views.NodesView();
            $('#main').html("");;
        }
    });



}());
