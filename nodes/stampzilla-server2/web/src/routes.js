import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Dashboard from './routes/dashboard';
import Debug from './routes/debug';
import Node from './routes/nodes/Node';
import Nodes from './routes/nodes';
import Security from './routes/security';

const Routes = () => (
  <Switch>
    <Route exact path="/" component={Dashboard} />
    <Route exact path="/nodes" component={Nodes} />
    <Route path="/nodes/:uuid" component={Node} />
    <Route exact path="/security" component={Security} />
    <Route path="/debug" component={Debug} />
  </Switch>
);

export default hot(module)(Routes);
