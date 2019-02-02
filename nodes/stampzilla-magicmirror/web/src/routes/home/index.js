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
          widgets.map(widget => {
            switch (widget.get('type')) {
              case 'clock':
                return <Clock />
              case 'forecast':
                return <Forecast />
              case 'devicelist':
                return <DeviceList devices={widget.get('devices')} />
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
