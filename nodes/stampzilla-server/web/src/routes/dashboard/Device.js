import React, { Component } from 'react';
import {
  mdiLightbulb,
  mdiPower,
  mdiPulse,
} from '@mdi/js';
import Icon from '@mdi/react';
import classnames from 'classnames';

import './device.scss';
import { write } from '../../components/Websocket';
import Trait from './Trait';

export const traitPriority = [
  'OnOff',
  'Brightness',
  'ColorSetting',
];

export const traitNames = {
  Brightness: 'Brightness',
  OnOff: 'Power',
  ColorSetting: 'Temperature',
};

export const traitStates = {
  Brightness: 'brightness',
  OnOff: 'on',
  ColorSetting: 'temperature',
};

const icons = {
  light: mdiLightbulb,
  switch: mdiPower,
  sensor: mdiPulse,
};

const guessType = (device) => {
  const traits = (device.get('traits') && device.get('traits').toJS()) || [];

  if (traits.length === 0) {
    return 'sensor';
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

class Device extends Component {
  onChange = (device, trait) => (value) => {
    const clone = device.toJS();
    clone.state[traitStates[trait]] = value;

    write({
      type: 'state-change',
      body: {
        [clone.id]: clone,
      },
    });
  }

  render() {
    const { device } = this.props;

    const sortedTraits = device.get('traits') && device.get('traits').sort((a, b) => {
      const prioA = traitPriority.findIndex(trait => trait === a);
      const prioB = traitPriority.findIndex(trait => trait === b);
      return prioA - prioB;
    });

    const primaryTrait = sortedTraits && sortedTraits.first();
    const secondaryTraits = sortedTraits && sortedTraits.shift();

    const type = device.get('type') || guessType(device);
    const icon = icons[type];

    return (
      <div className={classnames({
        'd-flex flex-column': true,
        offline: !device.get('online'),
      })}
      >
        <div className="d-flex align-items-center py-1">
          <div style={{ width: '1.5rem' }}>
            { icon &&
            <Icon
              path={icon}
              size={1}
            />
            }
          </div>
          <div className="flex-grow-1 mr-2">
            {device.get('name')}<br />
          </div>
          {primaryTrait &&
          <Trait
            trait={primaryTrait}
            device={device}
            state={traitStates[primaryTrait] && device.getIn(['state', traitStates[primaryTrait]])}
            onChange={this.onChange(device, primaryTrait)}
          />
          }
          {!primaryTrait &&
            <span>{JSON.stringify(device.get('state'))}</span>
          }
        </div>
        <div className="d-flex flex-column ml-4">
          {secondaryTraits && secondaryTraits.map(trait => (
            <div className="d-flex ml-3" key={trait}>
              <div className="mr-2">{traitNames[trait] || trait}</div>
              <div className="flex-grow-1 d-flex align-items-center">
                <Trait
                  trait={trait}
                  device={device}
                  state={traitStates[trait] && device.getIn(['state', traitStates[trait]])}
                  onChange={this.onChange(device, trait)}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }
}

export default Device;
