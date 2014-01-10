/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/

(function () {
    "use strict";
    //window.Stampzilla = window.Stampzilla || {Routers: {}, Collections: {}, Models: {}, Views: {}};
    Stampzilla.Views.NodesView = Backbone.View.extend({
        // the constructor
        initialize: function () {
        // model is passed through
        //this.notes.bind('reset', this.addAll, this);
        },
        // populate the html to the dom
        render: function () {
            this.$el.html(this.template());
            //this.addAll();
            return this;
        },

        addAll: function () {
            // clear out the container each time you render index
            this.$el.find('tbody').children().remove();
            _.each(this.notes.models, $.proxy(this, 'addOne'));
        },

        addOne: function (note) {
            var view = new Stampzilla.Views.NoteRowView({notes: this.notes, note: note});
            this.$el.find("tbody").append(view.render().el);
        }
    });
}());
