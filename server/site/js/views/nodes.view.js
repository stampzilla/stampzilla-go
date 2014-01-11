/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/

(function () {
    "use strict";
    //window.Stampzilla = window.Stampzilla || {Routers: {}, Collections: {}, Models: {}, Views: {}};
    Stampzilla.Views.NodesTable = Backbone.View.extend({
        initialize: function () {
            this.collection.bind('sync', this.render, this);
            this.collection.bind('add', this.addOne, this);
            this.$el.html(this.template());
        },
        // populate the html to the dom
        render: function () {
            this.addAll();
            return this;
        },
        addAll: function () {
            // clear out the container each time you render index
            this.$el.find('tbody').children().remove();
            _.each(this.collection.models, $.proxy(this, 'addOne'));
        },
        addOne: function (model) {
            var view = new Stampzilla.Views.NodesTableRow({model: model});
            this.$el.find("tbody").append(view.render().el);
        }
    });
    Stampzilla.Views.NodesTableRow = Backbone.View.extend({
        // the constructor
        tagName: "tr",
        // populate the html to the dom
        render: function () {
            this.$el.html(this.template(this.model.toJSON()));
            return this;
        },
        open: function(id){
            location.hash = "node/"+this.model.get('Id');
        },
        events: {
            "click" : "open",
        }

    });

    Stampzilla.Views.Node = Backbone.View.extend({
        initialize: function () {
            this.model.bind('change', this.render, this);
            this.$el.html(this.template());
        },
        // populate the html to the dom
        render: function () {


            _.each(this.model.get('Layout'), $.proxy(this, 'parseLayout'));
            
            return this;
        },
        parseLayout: function(layout){

            console.log(layout);
        }
    });

}());
