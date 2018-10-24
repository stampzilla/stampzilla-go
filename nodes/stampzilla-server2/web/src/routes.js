import React from "react";
import { Switch, Route } from 'react-router-dom';
import Nodes from './containers/Nodes';
import { hot } from 'react-hot-loader';

const Routes = () => (
    <Switch>
        <Route component={Nodes} />
    </Switch>
);

export default hot(module)(Routes);
