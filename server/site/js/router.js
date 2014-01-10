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
      "nodes/:id": "showNode",
    },

    initialize: function () {
    },

    index: function () {
      this.currentView = new Stampzilla.Views.NodesView();
      $('#main').html(this.currentView.render().el);
    },
    nodes: function () {
      this.currentView = new Stampzilla.Views.NodesView();
      $('#main').html("");;
    }
  });



}());
