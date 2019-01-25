import React, { Component } from 'react';
import { AlphaPicker } from 'react-color';
import SliderPointer from 'react-color/lib/components/slider/SliderPointer';
import ct from 'color-temperature';

import { uniqeId } from '../../helpers';

// import HueColorPicker from './HueColorPicker';

const temperatureToRgb = (temp) => {
  const color = ct.colorTemperature2rgb(temp);
  return `rgb(${color.red}, ${color.green}, ${color.blue})`;
};

const temperatureGradient = (start = 2000, end = 6000, steps = 10) => {
  const grad = [];
  for (let i = 0; i <= steps; i += 1) {
    grad.push(
      `${temperatureToRgb((((end - start) / steps) * i) + start)} ${(100 / steps) * i}%`,
    );
  }

  return grad.join(', ');
};


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
      case 'ColorSetting': {
        // <HueColorPicker />
        const start = 2000;
        const stop = 6500;
        return (<AlphaPicker
          style={{
              gradient: {
                background: `linear-gradient(to right, ${temperatureGradient(start, stop, 10)})`,
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
          onChange={({ rgb }) => onChange((rgb.a * (stop - start)) + start)}
          disabled={!device.get('online')}
        />
        );
      }
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
