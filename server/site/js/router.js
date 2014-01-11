/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/

window.Stampzilla = window.Stampzilla || {Routers: {}, Collections: {}, Models: {}, Views: {}};
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
    },

    index: function () {
        var collection = new Stampzilla.Collections.NodesCollection();
        collection.fetch();
        this.currentView = new Stampzilla.Views.NodesTable({collection:collection});
        $('#main').html(this.currentView.el);
    },
    showNode: function (id) {
        var model = new Stampzilla.Models.NodeModel({id:id});
        model.fetch();
        this.currentView = new Stampzilla.Views.Node({model:model});
        $('#main').html(this.currentView.el);
    },
    nodes: function () {
        //this.currentView = new Stampzilla.Views.NodesView();
        $('#main').html("");;
    }
  });



}());
