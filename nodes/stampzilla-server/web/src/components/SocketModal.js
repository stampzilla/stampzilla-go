import React, { Component } from 'react';

import FormModal from './FormModal';

const schema = {
  type: 'object',
  required: ['hostname', 'port'],
  properties: {
    hostname: { type: 'string', title: 'Hostname', default: location.hostname },
    port: { type: 'string', title: 'Port', default: location.port },
  },
};

class SocketModal extends Component {
  onChange = () => ({ formData }) => {
    this.props.onChange(formData);
    this.props.onClose();
  };

  render = () => {
    const { visible, onClose } = this.props;

    return (
      <FormModal
        title="Change websocket url"
        schema={schema}
        isOpen={visible}
        onClose={onClose}
        onSubmit={this.onChange()}
        submitButton="Change"
      />
    );
  };
}

export default SocketModal;
