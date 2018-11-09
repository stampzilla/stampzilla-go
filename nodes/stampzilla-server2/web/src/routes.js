import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Debug from './routes/debug';
import Node from './routes/nodes/Node';
import Nodes from './routes/nodes';

const Routes = () => (
  <Switch>
    <Route exact path="/nodes" component={Nodes} />
    <Route path="/nodes/:uuid" component={Node} />
    <Route path="/debug" component={Debug} />
  </Switch>
);

export default hot(module)(Routes);
