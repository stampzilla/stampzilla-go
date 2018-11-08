import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Debug from './routes/debug';
import Nodes from './routes/nodes';

const Routes = () => (
  <Switch>
    <Route path="/nodes" component={Nodes} />
    <Route path="/debug" component={Debug} />
  </Switch>
);

export default hot(module)(Routes);
