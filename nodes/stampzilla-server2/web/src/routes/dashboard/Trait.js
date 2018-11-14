import React, { Component } from 'react';
import { AlphaPicker } from 'react-color';
import SliderPointer from 'react-color/lib/components/slider/SliderPointer';

import { uniqeId } from '../../helpers';
import HueColorPicker from './HueColorPicker';

class Trait extends Component {
  renderTrait() {
    const { trait } = this.props;
    const id = uniqeId();

    switch (trait) {
      case 'OnOff': return (
        <span className="switch switch-sm">
          <input type="checkbox" className="switch" id={`switch-${id}`} />
          <label htmlFor={`switch-${id}`} className="mb-0" />
        </span>
      );
      case 'Brightness': return (
        <AlphaPicker
          style={{
            gradient: {
              background: 'linear-gradient(to right, #000 0%, #fff 100%)',
            },
          }}
          pointer={SliderPointer}
          height="12px"
          width="100%"
        />
      );
      case 'ColorSetting': return (
        <HueColorPicker />
      );
      default: return null;
    }
  }

  render() {
    return (
      <React.Fragment>
        {this.renderTrait()}
      </React.Fragment>
    );
  }
}

export default Trait;
