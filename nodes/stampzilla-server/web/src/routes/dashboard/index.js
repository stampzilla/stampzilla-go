import React, { Component } from 'react';
import { connect } from 'react-redux';

import Card from '../../components/Card';
import Device from './Device';

class Nodes extends Component {
  render() {
    const { devices, nodes } = this.props;

    const devicesByNode = devices.reduce((acc, device) => {
      const [node] = device.get('id').split('.', 2);
      if (acc[node] === undefined) {
        acc[node] = [];
      }
      acc[node].push(device);
      return acc;
    }, {});

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-4">
            {Object.keys(devicesByNode).map(nodeId => (
              <Card
                title={nodes.getIn([nodeId, 'name']) || `New node of type ${nodes.getIn([nodeId, 'type'])}`}
                bodyClassName="p-3"
                key={nodeId}
              >
                {devicesByNode[nodeId]
                  .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                  .map(device => (
                    <Device device={device} key={`${nodeId}.${device.get('id')}`} />
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
