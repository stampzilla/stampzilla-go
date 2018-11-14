import React, { Component } from 'react';
import {
  mdiLightbulb,
  mdiPower,
  mdiPulse,
} from '@mdi/js';
import Icon from '@mdi/react';

import Trait from './Trait';

const traitPriority = [
  'OnOff',
  'Brightness',
  'ColorSetting',
];

const traitNames = {
  Brightness: 'Brightness',
  OnOff: 'Power',
  ColorSetting: 'Color',
};

const traitStates = {
  Brightness: 'brightness',
  OnOff: 'on',
  ColorSetting: 'color',
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
  render() {
    const { device, state } = this.props;

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
      <div className="d-flex flex-column">
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
          <Trait trait={primaryTrait} device={device} state={traitStates[primaryTrait] && device.getIn(['state', traitStates[primaryTrait]])} />
          }
          {!primaryTrait &&
            <span>{JSON.stringify(state)}</span>
          }
        </div>
        <div className="d-flex flex-column ml-4">
          {secondaryTraits && secondaryTraits.map(trait => (
            <div className="d-flex ml-3" key={trait}>
              <div className="mr-2">{traitNames[trait] || trait}</div>
              <div className="flex-grow-1 d-flex align-items-center">
                <Trait trait={trait} device={device} state={traitStates[trait] && device.getIn(['state', traitStates[trait]])} />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }
}

export default Device;
