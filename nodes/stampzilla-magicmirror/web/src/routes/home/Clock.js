import React from 'react'
import moment from 'moment'

class Clock extends React.Component {
  constructor(props) {
    super(props)
    this.clock = React.createRef()
    this.seconds = React.createRef()
    this.date = React.createRef()
  }

  componentDidMount() {
    const updateClock = () => {
      this.clock.current.innerHTML = moment().format('HH:mm')
      this.seconds.current.innerHTML = moment().format('ss')
      this.date.current.innerHTML = moment().format('dddd - D MMMM')
    }
    updateClock()
    this.clockInterval = setInterval(updateClock, 1000)
  }

  componentWillUnmount() {
    clearTimeout(this.clockInterval)
  }

  render() {
    return (
      <React.Fragment>
        <div className="clock">
          <div ref={this.clock}>00:00</div>
          <div className="seconds" ref={this.seconds}>
            00
          </div>
        </div>
        <div className="date" ref={this.date}>
          -
        </div>
      </React.Fragment>
    )
  }
}

export default Clock
