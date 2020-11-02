import { request } from '../components/Websocket';
import reducer from '../ducks/persons';

const rules = (store) => (next) => async (action) => {
  if (action.type.startsWith('persons_') && !action.type.endsWith('_UPDATE')) {
    // Generate a updated state by running the reducer
    const updatedState = reducer(store.getState().get('persons'), action);

    // Try to save
    await request({
      type: 'update-persons',
      body: updatedState.get('list').toJS(),
    });

    // Jump to the next middleware
    return next(action);
  }

  return next(action);
};

export default rules;
