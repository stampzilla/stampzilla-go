import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Automation from './routes/automation';
import Dashboard from './routes/dashboard';
import Debug from './routes/debug';
import Persons from './routes/persons';
import Person from './routes/persons/Person';
import Nodes from './routes/nodes';
import Node from './routes/nodes/Node';
import Rule from './routes/automation/Rule';
import Schedule from './routes/automation/Schedule';
import Alerts from './routes/alerts';
import Security from './routes/security';
import Trigger from './routes/alerts/Trigger';
import Destination from './routes/alerts/Destination';
import Sender from './routes/alerts/Sender';
import { withBoudary } from './components/ErrorBoundary';

const Routes = () => (
  <Switch>
    <Route exact path="/" component={withBoudary(Dashboard)} />
    <Route exact path="/aut" component={withBoudary(Automation)} />
    <Route exact path="/aut/rule/create" component={withBoudary(Rule)} />
    <Route exact path="/aut/rule/:uuid" component={withBoudary(Rule)} />
    <Route
      exact
      path="/aut/schedule/create"
      component={withBoudary(Schedule)}
    />
    <Route exact path="/aut/schedule/:uuid" component={withBoudary(Schedule)} />
    <Route exact path="/persons" component={withBoudary(Persons)} />
    <Route path="/persons/:uuid" component={withBoudary(Person)} />
    <Route exact path="/nodes" component={withBoudary(Nodes)} />
    <Route path="/nodes/:uuid" component={withBoudary(Node)} />
    <Route exact path="/alerts" component={withBoudary(Alerts)} />
    <Route
      exact
      path="/alerts/triggers/create"
      component={withBoudary(Trigger)}
    />
    <Route
      exact
      path="/alerts/triggers/:uuid"
      component={withBoudary(Trigger)}
    />
    <Route
      exact
      path="/alerts/destinations/create"
      component={withBoudary(Destination)}
    />
    <Route
      exact
      path="/alerts/destinations/:uuid"
      component={withBoudary(Destination)}
    />
    <Route
      exact
      path="/alerts/senders/create"
      component={withBoudary(Sender)}
    />
    <Route exact path="/alerts/senders/:uuid" component={withBoudary(Sender)} />
    <Route exact path="/security" component={withBoudary(Security)} />
    <Route path="/debug" component={withBoudary(Debug)} />
  </Switch>
);

export default hot(module)(Routes);
