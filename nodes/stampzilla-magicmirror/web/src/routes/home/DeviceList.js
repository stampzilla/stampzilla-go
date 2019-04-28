import React from 'react'
import { connect } from 'react-redux'

class DeviceList extends React.PureComponent {
  render() {
    const { devices, state } = this.props

    return (
      <div className="device-list">
        {devices.map(device => {
          let value = JSON.stringify(
            state.getIn([device.get('device'), 'state', device.get('state')])
          )

          if (value && device.get('decimals')) {
            value *= Math.pow(10, device.get('decimals'))
            value = Math.round(value)
            value /= Math.pow(10, device.get('decimals'))
          }

          if (value && device.get('states')) {
            value = device.getIn(['states', value])
          }

          if (device.get('unit')) {
            value = `${value} ${device.get('unit')}`
          }

          return (
            <div
              className="device-list-item"
              key={`${device.get('device')}${device.get('state')}`}>
              <strong>{value}</strong>
              <small>{device.get('title')}</small>
            </div>
          )
        })}
      </div>
    )
  }
}

const mapStateToProps = state => ({
  state: state.getIn(['devices', 'list'])
})

export default connect(mapStateToProps)(DeviceList)
