import React, { Component } from 'react';

import { uniqeId } from '../../helpers';

class Trait extends Component {
  render() {
    const { trait } = this.props;
    const id = uniqeId();

    switch(trait) {
      case 'OnOff': return (
        <span className="switch switch-sm">
          <input type="checkbox" className="switch" id={`switch-${id}`} />
          <label htmlFor={`switch-${id}`} className="mb-0"/>
        </span>
      );
      case 'Brightness': return (
        <input type="range" className="form-control-range" />
      );
      default: return null;
    }
  }
}

export default Trait;
