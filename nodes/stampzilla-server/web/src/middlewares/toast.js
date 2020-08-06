import React from 'react';
import { toast } from 'react-toastify';

const toastify = (store) => (next) => async (action) => {
  try {
    return await next(action);
  } catch (ex) {
    toast.error(
      <>
        <strong>Save failed:</strong>
        {' '}
        {ex}
      </>,
    );

    throw ex;
  }
};

export default toastify;
