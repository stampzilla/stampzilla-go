import React, { Component } from 'react';
import { AlphaPicker } from 'react-color';
import SliderPointer from 'react-color/lib/components/slider/SliderPointer';

import { uniqeId } from '../../helpers';
import HueColorPicker from './HueColorPicker';

class Trait extends Component {
  renderTrait() {
    const {
      trait, state, onChange, device,
    } = this.props;
    const id = uniqeId();

    switch (trait) {
      case 'OnOff': return (
        <span className="switch switch-sm">
          <input
            type="checkbox"
            className="switch"
            id={`switch-${id}`}
            checked={!!state}
            onChange={event => onChange(event.target.checked)}
            disabled={!device.get('online')}
          />
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
          color={{
              r: 0,
              g: 0,
              b: 0,
              a: state,
          }}
          onChange={({ rgb }) => onChange(rgb.a)}
          disabled={!device.get('online')}
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
