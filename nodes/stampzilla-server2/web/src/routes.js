import React from "react";
import { Switch, Route } from 'react-router-dom';
import Debug from './containers/Debug';
import { hot } from 'react-hot-loader';

const Routes = () => (
    <Switch>
        <Route component={Debug} />
    </Switch>
);

export default hot(module)(Routes);
