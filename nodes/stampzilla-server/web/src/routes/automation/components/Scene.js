import React from 'react';
import { connect } from 'react-redux';
import diff from 'immutable-diff';
import classnames from 'classnames';

import './Scene.scss';
import Trait from '../../dashboard/Trait';
import { traitPriority, traitNames, traitStates } from '../../dashboard/Device';

// List of traits that is used to filter out useable devices
const traits = ['OnOff', 'Brightness', 'ColorSetting'];

const checkBox = ({
  device, trait, checked, onChange,
}) => (
  <div className="checkbox custom-control custom-checkbox ">
    <input
      type="checkbox"
      className="custom-control-input"
      id={`${device.get('id')}${trait}`}
      checked={checked}
      onChange={event => onChange(event.target.checked)}
    />
    <label
      className="custom-control-label"
      htmlFor={`${device.get('id')}${trait}`}
    />
  </div>
);

class Scene extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      recording: false,
      states: {},
    };
  }

  componentWillReceiveProps(props) {
    const { devices } = this.props;

    const { recording } = this.state;
    const states = (this.state && this.state.states) || {};

    const useableDevices = devices.reduce((acc, dev) => {
      if (
        traits.find(trait => (dev.get('traits') || []).indexOf(trait) !== -1)
      ) {
        acc.push(dev.get('id'));
      }
      return acc;
    }, []);

    if (recording) {
      const changes = diff(props.devices, this.props.devices);

      changes.forEach((change) => {
        const [device, type, key] = change.get('path').toJS();
        if (type !== 'state') {
          return; // No state change
        }
        if (useableDevices.indexOf(device) === -1) {
          return; // Not a useable device
        }

        states[device] = {
          ...states[device],
          [key]: props.devices.getIn(change.get('path')),
        };
      });

      this.setState({
        states,
      });
    }
  }

  onToggleDevice = device => (checked) => {
    const states = (this.state && this.state.states) || {};
    if (checked) {
      states[device.get('id')] = states[device.get('id')] || {};
    } else {
      delete states[device.get('id')];
    }
    this.setState({
      states,
    });
  };
  onToggleTrait = (device, trait) => (checked) => {
    const states = (this.state && this.state.states) || {};
    if (checked) {
      states[device.get('id')] = states[device.get('id')] || {};
      states[device.get('id')][traitStates[trait]] = device.getIn([
        'state',
        traitStates[trait],
      ]);
    } else {
      delete states[device.get('id')][traitStates[trait]];
    }
    this.setState({
      states,
    });
  };

  onChangeTrait = (device, trait) => (value) => {
    const states = (this.state && this.state.states) || {};

    states[device.get('id')] = states[device.get('id')] || {};
    states[device.get('id')][traitStates[trait]] = value;

    this.setState({
      states,
    });
  };

  render() {
    const { devices } = this.props;
    const { recording, states } = this.state;

    const useableDevices = devices
      .reduce((acc, dev) => {
        if (
          traits.find(trait => (dev.get('traits') || []).indexOf(trait) !== -1)
        ) {
          acc.push(dev);
        }
        return acc;
      }, [])
      .sort((a, b) => a.get('name').localeCompare(b.get('name')));

    return (
      <div className="saved-state-builder">
        <div className="menu mb-1">
          <button
            className={recording ? 'btn btn-danger' : 'btn btn-secondary'}
            onClick={() => this.setState({ recording: !recording })}
            type="button"
          >
            {recording ? 'Recording' : 'Record'}
          </button>
          <button className="btn btn-secondary ml-1 hide" type="button">
            Select all lights
          </button>
        </div>
        <div className="devices">
          {useableDevices.map((dev) => {
            const sortedTraits =
              dev.get('traits') &&
              dev.get('traits').sort((a, b) => {
                const prioA = traitPriority.findIndex(trait => trait === a);
                const prioB = traitPriority.findIndex(trait => trait === b);
                return prioA - prioB;
              });
            const selected =
              states && Object.keys(states).indexOf(dev.get('id')) !== -1;

            return (
              <div
                key={dev.get('id')}
                className={classnames({
                  selected,
                })}
              >
                <div className="d-flex mt-2">
                  {checkBox({
                    device: dev,
                    checked: selected,
                    onChange: this.onToggleDevice(dev),
                  })}
                  {dev.get('name')}
                </div>

                {sortedTraits &&
                  sortedTraits.map(trait => (
                    <div className="d-flex ml-3 trait" key={trait}>
                      {checkBox({
                        device: dev,
                        trait: traitStates[trait],
                        checked:
                          states &&
                          states[dev.get('id')] &&
                          states[dev.get('id')][traitStates[trait]] !==
                            undefined,
                        onChange: this.onToggleTrait(dev, trait),
                      })}
                      <div className="mr-2">{traitNames[trait] || trait}</div>
                      <div className="flex-grow-1 d-flex align-items-center justify-content-end">
                        <Trait
                          trait={trait}
                          device={dev}
                          state={
                            states &&
                            states[dev.get('id')] &&
                            states[dev.get('id')][traitStates[trait]] !==
                              undefined
                              ? states[dev.get('id')][traitStates[trait]]
                              : traitStates[trait] &&
                                dev.getIn(['state', traitStates[trait]])
                          }
                          onChange={this.onChangeTrait(dev, trait)}
                        />
                      </div>
                    </div>
                  ))}
              </div>
            );
          })}
        </div>
      </div>
    );
  }
}

const mapStateToProps = state => ({
  devices: state.getIn(['devices', 'list']),
  states: state.getIn(['savedstates', 'list']),
});

export default connect(mapStateToProps)(Scene);
