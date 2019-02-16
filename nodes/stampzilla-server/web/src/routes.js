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

const Routes = () => (
  <Switch>
    <Route exact path="/" component={Dashboard} />
    <Route exact path="/aut" component={Automation} />
    <Route exact path="/aut/rule/create" component={Rule} />
    <Route exact path="/aut/rule/:uuid" component={Rule} />
    <Route exact path="/aut/schedule/create" component={Schedule} />
    <Route exact path="/aut/schedule/:uuid" component={Schedule} />
    <Route exact path="/nodes" component={Nodes} />
    <Route path="/nodes/:uuid" component={Node} />
    <Route exact path="/security" component={Security} />
    <Route path="/debug" component={Debug} />
  </Switch>
);

export default hot(module)(Routes);
