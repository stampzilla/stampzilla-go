/*global Stampzilla:true, _:true, jQuery:true, Backbone:true, JST:true, $:true*/
(function () {
    "use strict";
    Stampzilla.Views.NodesTable = Backbone.View.extend({
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
            this.listenTo(this.model,'change', this.render, this);
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

            //if( layout.Type == "switch" && layout.Action == "toggle"){

            //}

            //TODO change Using from Devices to Devices[Type=!dimmable] or add filter element
            //
            var states = this.parseStates(layout.Using,this.model.get('State').get(layout.Using));

            console.log(states);

            //loop each state and create NodeActionRow view
            _.each(states, function(state){
                this.ActionSubviews[state.Id] = new Stampzilla.Views.NodeActionRow(state,this.model.get('State'));
                this.$el.find('#nodeList').append(this.ActionSubviews[state.Id].render().el);
            },this);

        },
        parseStates: function(key,states){
            var ret = [];
            _.each(states, function(state){
                state.Actions = {};
                _.each(state.Features, function(f){
                    state.Actions[f] = this.getAction(f);
                },this);
                ret.push(state);
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
    Stampzilla.Views.NodeActionRow = Backbone.View.extend({
        tagName: 'li',
        initialize: function(state,model) {
            this.state = state;
            this.model = model;
            this.actionViews = []
        },
        render: function () {
            console.log(this.state);


            this.$el.html(this.template(this.state));

            _.each(this.state.Actions, function(action){
                //this.$el.find('.actions').append(this.generateActionElement(action));
                this.actionViews[action.Name] = new Stampzilla.Views.NodeActionDiv(action,this);
                this.$el.find('.actions').append(this.actionViews[action.Name].render().el);
            },this);

            return this;
        },

        generateActionElement: function(action){
            var actionElement = $("<div>",{
                id:action.Id,
                "class": 'test',
                text:action.Name,
            });
            return actionElement;
        },

    });
    Stampzilla.Views.NodeActionDiv = Backbone.View.extend({
        tagName: 'div',
        className:'test',
        initialize: function(action,parentView) {
            this.action = action;
            this.state = parentView.state;
            this.model = parentView.model;
        },
        render: function () {
            //var actionElement = $("<div>",{
                //id:action.Id,
                //"class": 'test',
                //text:action.Name,
            //});
            this.$el.html(this.action.Name);
            return this;
        },
        runAction: function(e){
            var args = [],tmp;

            _.each(this.action.Arguments, function(arg){
                tmp = arg.split('.');
                args.push(this.state[tmp[1]])
                args.push(this.action.Id);

            },this);

            this.model.clear({silent:true});
            this.model.save({To:this.model.node.get('Id'),args:args});
            // this posts to /api/Tellstick/state
            // TODO
            // make sure we get new State back from sever here for our device (Tellstick in this case)
            // 
            //
        },
        events: {
            "click" : "runAction",
        }
    });

}());
