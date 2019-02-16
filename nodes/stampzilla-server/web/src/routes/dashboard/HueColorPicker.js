import React, { PureComponent } from 'react';
import { CustomPicker } from 'react-color';
import { Hue } from 'react-color/lib/components/common';
import SliderPointer from 'react-color/lib/components/slider/SliderPointer';

class HueColorPicker extends PureComponent {
  render() {
    return (
      <div style={{ height: '12px', position: 'relative', width: '100%' }}>
        <Hue
          {...this.props}
          style={{ radius: '2px' }}
          direction="horizontal"
          pointer={SliderPointer}
        />
      </div>
    );
  }
}

export default CustomPicker(HueColorPicker);
