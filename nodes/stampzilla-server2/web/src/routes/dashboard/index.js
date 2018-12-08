import React, { Component } from 'react';
import { connect } from 'react-redux';

import Card from '../../components/Card';
import Device from './Device';

class Nodes extends Component {
  render() {
    const { devices, nodes } = this.props;

    const devicesByNode = devices.reduce((acc, device) => {
      if (acc[device.get('node')] === undefined) {
        acc[device.get('node')] = [];
      }
      acc[device.get('node')].push(device);
      return acc;
    }, {});

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-4">
            {Object.keys(devicesByNode).map(nodeId => (
              <Card
                title={nodes.getIn([nodeId, 'name'])}
                bodyClassName="p-3"
              >
                {devicesByNode[nodeId]
                  .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                  .map(device => (
                    <Device device={device} />
                ))}
              </Card>
              ))}
          </div>
        </div>
      </React.Fragment>
    );
  }
}

const mapToProps = state => ({
  devices: state.getIn(['devices', 'list']),
  nodes: state.getIn(['nodes', 'list']),
});

export default connect(mapToProps)(Nodes);
