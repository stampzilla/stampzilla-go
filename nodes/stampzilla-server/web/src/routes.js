import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Automation from './routes/automation';
import Dashboard from './routes/dashboard';
import Debug from './routes/debug';
import Node from './routes/nodes/Node';
import Nodes from './routes/nodes';
import Rule from './routes/automation/Rule';
import Schedule from './routes/automation/Schedule';
import Security from './routes/security';
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
    <Route exact path="/nodes" component={withBoudary(Nodes)} />
    <Route path="/nodes/:uuid" component={withBoudary(Node)} />
    <Route exact path="/security" component={withBoudary(Security)} />
    <Route path="/debug" component={withBoudary(Debug)} />
  </Switch>
);

export default hot(module)(Routes);
