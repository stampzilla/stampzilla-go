import { write } from '../components/Websocket';

const senders = (store) => (next) => (action) => {
  const prev = store.getState().getIn(['destinations', 'list']);
  const result = next(action);
  const after = store.getState().getIn(['destinations', 'list']);

  if (!after.equals(prev) && action.type !== 'destinations_UPDATE') {
    write({
      type: 'update-destinations',
      body: after.toJS(),
    });
  }
  return result;
};

export default senders;
