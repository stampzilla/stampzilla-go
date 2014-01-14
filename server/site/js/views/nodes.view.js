/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
(function () {
    "use strict";
    var NodesTable = Backbone.View.extend({
        initialize: function () {
            //fetch in router triggers this sync
            this.listenTo(this.collection,'sync reset', this.render, this);
            this.listenTo(this.collection,'add', this.addOne, this);
        },
        // populate the html to the dom
        render: function () {
            this.$el.html(this.template());
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
    Stampzilla.registerView('NodesTable',NodesTable);

    var NodesTableRow = Backbone.View.extend({
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

    Stampzilla.registerView('NodesTableRow',NodesTableRow);


    var Node = Backbone.View.extend({
        initialize: function () {
            this.listenTo(this.model,'change', this.render, this);
            this.listenTo(this.model.get('State'),'change', this.render, this);
            this.ActionSubviews = [];
        },
        // populate the html to the dom
        render: function () {
            this.$el.html(this.template());
            _.each(this.model.get('Layout'), $.proxy(this, 'parseLayout'));
            return this;
        },
        parseLayout: function(layout){
            var data = {};

            //TODO change Using from Devices to Devices[Type=!dimmable] or add filter element
            //
            var states = this.parseStates(layout,this.model.get('State').get(layout.Using));

            //loop each state and create NodeActionRow view
            _.each(states, function(state){
                this.ActionSubviews[state.Id] = new Stampzilla.Views.NodeActionRow(state,this.model.get('State'));
                this.$el.find('#nodeList').append(this.ActionSubviews[state.Id].render().el);
            },this);

        },
        parseStates: function(layout,states){
            var ret = [];
            _.each(states, function(state){
            
                _.each(layout.Filter, function(line){
                    if($.inArray(line, state.Features) != -1){
                        state.Action = this.getAction(layout.Action);
                        ret.push(state);
                    }
                },this);


            },this);
            return ret;
        },
        getAction: function(id){
            var ret = undefined;

            var actions = this.model.get('Actions');
            _.each(actions, function(row){
                if(row.Id === id){
                    ret = row;
                }
            });

            return ret;
        }

    });
    Stampzilla.registerView('Node',Node);

    var NodeActionRow = Backbone.View.extend({
        tagName: 'li',
        initialize: function(state,model) {
            this.state = state;
            this.model = model;
            this.actionViews = []
        },
        render: function () {


            this.$el.html(this.template(this.state));

            this.actionViews[this.state.Action] = new Stampzilla.Views.NodeActionDiv(this.state.Action,this);
            this.$el.find('.actions').append(this.actionViews[this.state.Action].render().el);

            return this;
        },

    });
    Stampzilla.registerView('NodeActionRow',NodeActionRow);

    var NodeActionDiv = Backbone.View.extend({
        tagName: 'button',
        className:'btn btn-default',
        initialize: function(action,parentView) {
            this.action = action;
            this.state = parentView.state;
            this.model = parentView.model;
        },
        render: function () {

            if(this.state.State == "true"){
                this.$el.addClass('btn-primary').removeClass('btn-default');
            } else{
                this.$el.addClass('btn-default').removeClass('btn-primary');
            }

            this.$el.html(this.action.Name+this.state.State);
            return this;
        },
        runAction: function(e){
            var args = [],tmp;

            _.each(this.action.Arguments, function(arg){
                tmp = arg.split('.');
                args.push(this.state[tmp[1]])
            },this);

            this.model.clear({silent:true});
            this.model.save({Cmd:this.action.Id,Args:args},{wait:true});
            // this posts to /api/Tellstick/state
        },
        events: {
            "click" : "runAction",
        }
    });
    Stampzilla.registerView('NodeActionDiv',NodeActionDiv);

}());
