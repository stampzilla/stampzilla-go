import { write } from '../components/Websocket';

const rules = store => next => (action) => {
  const prev = store.getState().getIn(['rules', 'list']);
  const result = next(action);
  const after = store.getState().getIn(['rules', 'list']);

  if (!after.equals(prev) && action.type !== 'rules_UPDATE') {
    write({
      type: 'update-rules',
      body: after.toJS(),
    });
  }
  return result;
};

export default rules;
