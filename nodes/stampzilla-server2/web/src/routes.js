import { Route, Switch } from 'react-router-dom';
import { hot } from 'react-hot-loader';
import React from 'react';

import Debug from './containers/Debug';

const Routes = () => (
  <Switch>
    <Route component={Debug} />
  </Switch>
);

export default hot(module)(Routes);
