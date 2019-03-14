import React, { Component } from 'react';
import {
  mdiLightbulb,
  mdiPower,
  mdiPulse,
  mdiVolumeHigh,
} from '@mdi/js';
import Form from 'react-jsonschema-form';
import Icon from '@mdi/react';
import Modal from 'react-modal';
import ReactJson from 'react-json-view';
import classnames from 'classnames';

import './device.scss';
import { write } from '../../components/Websocket';
import CustomCheckbox from '../../components/CustomCheckbox';
import Trait from './Trait';

export const traitPriority = ['OnOff', 'Brightness', 'ColorSetting'];

export const traitNames = {
  Brightness: 'Brightness',
  ColorSetting: 'Temperature',
  OnOff: 'Power',
  Volume: 'Volume',
};

export const traitStates = {
  Brightness: 'brightness',
  ColorSetting: 'temperature',
  OnOff: 'on',
  Volume: 'volume',
};

const icons = {
  light: mdiLightbulb,
  switch: mdiPower,
  sensor: mdiPulse,
  audio: mdiVolumeHigh,
};

const guessType = (device) => {
  const traits = (device.get('traits') && device.get('traits').toJS()) || [];

  if (traits.length === 0) {
    return 'sensor';
  }

  if (traits.indexOf('Volume') !== -1) {
    return 'audio';
  }
  if (traits.indexOf('Brightness') !== -1) {
    return 'light';
  }
  if (traits.indexOf('ColorSetting') !== -1) {
    return 'light';
  }
  if (traits.indexOf('OnOff') !== -1) {
    return 'switch';
  }

  return null;
};

const schema = {
  type: 'object',
  required: [],
  properties: {
    alias: {
      type: 'string',
      title: 'Alias',
    },
    /* config: {
      type: 'string',
      title: 'Config',
    }, */
  },
};
const uiSchema = {
  config: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 15,
    },
  },
};

class Device extends Component {
  state = {
    isValid: true,
    formData: {},
    modalIsOpen: false,
  };

  componentWillReceiveProps(props) {
    const { modalIsOpen } = this.state;
    if (!modalIsOpen) {
      this.setState({
        formData: {
          alias: props.device.get('alias'),
        },
      });
    }
  }

  onChange = (device, trait) => (value) => {
    const clone = device.toJS();
    clone.state[traitStates[trait]] = value;

    write({
      type: 'state-change',
      body: {
        [clone.id]: clone,
      },
    });
  };

  onModalChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  };

  onSubmit = () => ({ formData }) => {
    const { device } = this.props;
    // const node = nodes.find(n => n.get('uuid') === match.params.uuid);
    write({
      type: 'setup-device',
      body: {
        id: device.get('id'),
        ...formData,
        config: formData.config ? JSON.parse(formData.config) : undefined,
      },
    });
    this.setState({ modalIsOpen: false });
  };

  openModal = () => () => {
    this.setState({ modalIsOpen: true });
  };

  closeModal = () => () => {
    this.setState({ modalIsOpen: false });
  };

  render() {
    const { modalIsOpen, isValid, formData } = this.state;
    const { device } = this.props;

    const sortedTraits = device.get('traits')
      && device.get('traits').sort((a, b) => {
        const prioA = traitPriority.findIndex(trait => trait === a);
        const prioB = traitPriority.findIndex(trait => trait === b);
        if (prioA < 0 || prioB < 0) {
          return prioB - prioA;
        }
        return prioA - prioB;
      });

    const primaryTrait = sortedTraits && sortedTraits.first();
    const secondaryTraits = sortedTraits && sortedTraits.shift();

    const type = device.get('type') || guessType(device);
    const icon = icons[type];

    /* eslint-disable react/no-array-index-key */
    return (
      <div
        className={classnames({
          'd-flex flex-column device': true,
          offline: !device.get('online'),
        })}
      >
        <div className="d-flex align-items-center py-1">
          <div
            style={{ width: '1.5rem' }}
            onClick={this.openModal()}
            role="button"
            tabIndex={0}
          >
            {icon && <Icon path={icon} size={1} />}
          </div>
          <div
            className="flex-grow-1 mr-2"
            onClick={this.openModal()}
            role="button"
            tabIndex={0}
          >
            {device.get('alias') || device.get('name')}
            <br />
          </div>
          {primaryTrait && (
            <Trait
              trait={primaryTrait}
              device={device}
              state={device.getIn([
                'state',
                traitStates[primaryTrait] || primaryTrait.toLowerCase(),
              ])}
              onChange={this.onChange(device, primaryTrait)}
            />
          )}
          {!primaryTrait
            && device.get('state').size
            && device.get('state').size === 1
            && device
              .get('state')
              .map((v, k) => (
                <div className="d-flex ml-3" key={k}>
                  <div className="mr-2">{k}</div>
                  <div className="flex-grow-1 text-right">
                    {JSON.stringify(v)}
                  </div>
                </div>
              ))
              .valueSeq()
              .toArray()}
        </div>
        <div className="d-flex flex-column ml-4">
          {secondaryTraits
            && secondaryTraits.map(trait => (
              <div className="d-flex ml-3" key={trait}>
                <div className="mr-2">{traitNames[trait] || trait}</div>
                <div className="flex-grow-1 d-flex align-items-center">
                  <Trait
                    trait={trait}
                    device={device}
                    state={device.getIn([
                      'state',
                      traitStates[trait] || trait.toLowerCase(),
                    ])}
                    onChange={this.onChange(device, trait)}
                  />
                </div>
              </div>
            ))}
          {!primaryTrait
            && device.get('state').size
            && device.get('state').size > 1
            && device
              .get('state')
              .map((v, k) => (
                <div className="d-flex ml-3" key={k}>
                  <div className="mr-2">{k}</div>
                  <div className="flex-grow-1 text-right">
                    {JSON.stringify(v)}
                  </div>
                </div>
              ))
              .valueSeq()
              .toArray()}
        </div>

        <Modal
          className="Modal__Bootstrap modal-dialog saved-state-modal"
          closeTimeoutMS={150}
          isOpen={modalIsOpen || false}
          onRequestClose={this.closeModal()}
        >
          <div className="modal-content">
            <div className="modal-header">
              <h5 className="modal-title">Edit device</h5>
            </div>
            <div className="modal-body">
              <ReactJson
                name={false}
                theme="solarized"
                src={device && device.toJS()}
                displayDataTypes={false}
                displayObjectSize={false}
              />
              <hr />
              <Form
                schema={schema}
                uiSchema={uiSchema}
                showErrorList={false}
                liveValidate
                onChange={this.onModalChange()}
                formData={formData}
                onSubmit={this.onSubmit()}
                widgets={{
                  CheckboxWidget: CustomCheckbox,
                }}
              >
                <button
                  ref={(btn) => {
                    this.submitButton = btn;
                  }}
                  style={{ display: 'none' }}
                  type="submit"
                />
              </Form>
            </div>
            <div className="modal-footer">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={this.closeModal()}
              >
                Close
              </button>
              {isValid && (
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={() => this.submitButton.click()}
                >
                  Save
                </button>
              )}
            </div>
          </div>
        </Modal>
      </div>
    );
  }
}

export default Device;
