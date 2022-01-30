import React, { Component } from 'react';
import { AlphaPicker } from 'react-color';
import SliderPointer from 'react-color/lib/components/slider/SliderPointer';
import Switch from 'react-switch';
import ct from 'color-temperature';

import { uniqeId } from '../../../helpers';

// import HueColorPicker from './HueColorPicker';

const temperatureToRgb = (temp) => {
  const color = ct.colorTemperature2rgb(temp);
  return `rgb(${color.red}, ${color.green}, ${color.blue})`;
};

const temperatureGradient = (start = 2000, end = 6000, steps = 10) => {
  const grad = [];
  for (let i = 0; i <= steps; i += 1) {
    const temp = ((end - start) / steps) * i;
    grad.push(`${temperatureToRgb(temp + start)} ${(100 / steps) * i}%`);
  }

  return grad.join(', ');
};

class StateEditor extends Component {
  renderStateEditor() {
    const {
      state, onChange, device, arrayKey,
    } = this.props;
    const id = uniqeId();
    const type = typeof state;

    switch (type) {
      case 'boolean':
        return (
          <label htmlFor={`switch-${id}`} className="mb-0 d-flex">
            <Switch
              checked={!!state || false}
              onChange={(checked) => onChange(device, arrayKey, checked)}
              id={`switch-${id}`}
            />
          </label>
        );

      case 'Brightness':
      case 'Volume':
        return (
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
              a: state || 0,
            }}
            onChange={({ rgb }) => onChange(rgb.a)}
            disabled={!device.get('online') || !onChange}
          />
        );
      default:
        return (
          <input
            type="text"
            style={{width:'100%'}}
            value={state}
            onChange={(event) => onChange(device, arrayKey, event.target.value)}
          />
        );
    }
  }

  render() {
    return <>{this.renderStateEditor()}</>;
  }
}

export default StateEditor;
