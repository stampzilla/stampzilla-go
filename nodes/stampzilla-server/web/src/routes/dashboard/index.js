import React, { Component } from 'react';
import { Tab, Tabs } from 'react-bootstrap';
import { connect } from 'react-redux';

import Card from '../../components/Card';
import Device from './Device';

class Dashboard extends Component {
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

    const devicesByRoom = devices.reduce((acc, device) => {
      const node = device.getIn(['labels', 'room']);
      if (!node) {
        return acc;
      }

      if (acc[node] === undefined) {
        acc[node] = [];
      }
      acc[node].push(device);
      return acc;
    }, {});

    const cloudEnabledDevices = devices.reduce((acc, device) => {
      if (!device.getIn(['labels', 'cloud'])) {
        return acc;
      }

      const [node] = device.get('id').split('.', 2);
      if (acc[node] === undefined) {
        acc[node] = [];
      }
      acc[node].push(device);
      return acc;
    }, {});

    return (
      <Tabs>
        {Object.keys(devicesByRoom).length > 0 && (
          <Tab eventKey="byroom" title="Rooms">
            <div
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                alignContent: 'stretch',
              }}
            >
              {Object.keys(devicesByRoom)
                .sort((a, b) => a.localeCompare(b))
                .map((room) => (
                  <div
                    style={{
                      width: '400px',
                      maxWidth: '100%',
                      padding: '10px',
                      display: 'flex',
                    }}
                    key={room}
                  >
                    <Card
                      title={room}
                      bodyClassName="p-3"
                      className="mb-0 flex-grow-1"
                    >
                      {devicesByRoom[room]
                        .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                        .map((device) => (
                          <Device
                            device={device}
                            key={`${room}.${device.get('id')}`}
                          />
                        ))}
                    </Card>
                  </div>
                ))}
            </div>
          </Tab>
        )}

        <Tab eventKey="bynode" title="Nodes">
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              alignContent: 'stretch',
            }}
          >
            {Object.keys(devicesByNode).map((nodeId) => (
              <div
                style={{
                  width: '400px',
                  maxWidth: '100%',
                  padding: '10px',
                  display: 'flex',
                }}
                key={nodeId}
              >
                <Card
                  title={
                    nodes.getIn([nodeId, 'name'])
                    || `New node of type ${nodes.getIn([nodeId, 'type'])}`
                  }
                  bodyClassName="p-3"
                  className="mb-0 flex-grow-1"
                >
                  {devicesByNode[nodeId]
                    .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                    .map((device) => (
                      <Device
                        device={device}
                        key={`${nodeId}.${device.get('id')}`}
                      />
                    ))}
                </Card>
              </div>
            ))}
          </div>
        </Tab>

        {Object.keys(cloudEnabledDevices).length > 0 && (
          <Tab eventKey="bycloud" title="Cloud enabled">
            <div
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                alignContent: 'stretch',
              }}
            >
              {Object.keys(cloudEnabledDevices).map((node) => (
                <div
                  style={{
                    width: '400px',
                    maxWidth: '100%',
                    padding: '10px',
                    display: 'flex',
                  }}
                  key={node}
                >
                  <Card
                    title={
                      nodes.getIn([node, 'name'])
                      || `New node of type ${nodes.getIn([node, 'type'])}`
                    }
                    bodyClassName="p-3"
                    className="mb-0 flex-grow-1"
                  >
                    {cloudEnabledDevices[node]
                      .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                      .map((device) => (
                        <Device
                          device={device}
                          key={`${node}.${device.get('id')}`}
                        />
                      ))}
                  </Card>
                </div>
              ))}
            </div>
          </Tab>
        )}
      </Tabs>
    );
  }
}

const mapToProps = (state) => ({
  devices: state.getIn(['devices', 'list']),
  nodes: state.getIn(['nodes', 'list']),
});

export default connect(mapToProps)(Dashboard);
