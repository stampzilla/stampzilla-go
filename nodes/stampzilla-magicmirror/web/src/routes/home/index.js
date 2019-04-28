import React from 'react'
import { connect } from 'react-redux'

import Clock from './Clock'
import Forecast from './Forecast'
import DeviceList from './DeviceList'

class Home extends React.PureComponent {
  render() {
    const { widgets } = this.props
    return (
      <div>
        {widgets &&
          widgets.map((widget, index) => {
            switch (widget.get('type')) {
              case 'clock':
                return <Clock key={`clock${index}`} />
              case 'forecast':
                return <Forecast key={`forecast${index}`} config={widget} />
              case 'devicelist':
                return (
                  <DeviceList
                    key={`devicelist${index}`}
                    devices={widget.get('devices')}
                  />
                )
              default:
                return null
            }
          })}
      </div>
    )
  }
}

const mapStateToProps = state => ({
  widgets: state.getIn(['config', 'widgets'])
})

export default connect(mapStateToProps)(Home)
