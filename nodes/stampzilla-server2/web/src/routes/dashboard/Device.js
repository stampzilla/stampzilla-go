import React, { Component } from 'react';

import Trait from './Trait';

const traitPriority = [
  'OnOff',
  'Brightness',
  'ColorSetting',
];

const traitNames = {
  'Brightness': 'Brightness',
  'OnOff': 'Power',
  'ColorSetting': 'Color',
};

const traitStates = {
  'Brightness': 'brightness',
  'OnOff': 'on',
  'ColorSetting': 'color',
};

class Device extends Component {
  render() {
    const { device } = this.props;

    const sortedTraits = device.get('traits') && device.get('traits').sort((a, b) => {
      const prioA = traitPriority.findIndex(trait => trait === a);
      const prioB = traitPriority.findIndex(trait => trait === b);
      return prioA-prioB;
    });

    const primaryTrait = sortedTraits && sortedTraits.first();
    const secondaryTraits = sortedTraits && sortedTraits.shift();

    return (
      <div className="d-flex flex-column p-1">
        <div className="d-flex">
          <div className="flex-grow-1 mr-2">
          {device.get('name')}<br />
          </div>
          {primaryTrait && 
          <Trait trait={primaryTrait} device={device}/>
          }
        </div>
        <div className="d-flex flex-column ml-4">
          {secondaryTraits && secondaryTraits.map(trait => (
            <div className="d-flex" key={trait}>
              <div className="flex-grow-1 mr-2">{traitNames[trait] || trait}</div>
              <Trait trait={trait} device={device}/>
            </div>
          ))}
        </div>
      </div>
    );
  }
}

export default Device;
