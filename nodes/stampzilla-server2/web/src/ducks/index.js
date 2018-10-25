import { combineReducers } from 'redux-immutable';

import connection from './connection';

const rootReducer = combineReducers({
  connection
})

export default rootReducer;
