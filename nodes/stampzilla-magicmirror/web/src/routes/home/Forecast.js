import React from 'react'
import { connect } from 'react-redux'
import moment from 'moment'
import Moment from 'react-moment'

import { loadForecast } from '../../ducks/forecast'

const calendarStrings = {
  lastDay: '[Yesterday]',
  sameDay: '[Today]',
  nextDay: '[Tomorrow]',
  lastWeek: '[last] dddd [at] LT',
  nextWeek: 'dddd',
  sameElse: 'L'
}

const iconMap = {
  Sun: 'sunny',
  LightCloud: 'cloudy',
  PartlyCloud: 'cloudy',
  Cloud: 'cloudy',
  LightRainSun: 'rainy',
  LightRainThunderSun: 'rainy',
  SleetSun: 'rainy',
  SnowSun: 'snowy',
  LightRain: 'rainy',
  Rain: 'rainy',
  RainThunder: 'stormy',
  Sleet: 'rainy',
  Snow: 'rainy',
  SnowThunder: 'stormy',
  Fog: 'cloudy',
  SleetSunThunder: 'stormy',
  SnowSunThunder: 'stormy',
  LightRainThunder: 'stormy',
  SleetThunder: 'stormy',
  DrizzleThunderSun: 'stormy',
  RainThunderSun: 'stormy',
  LightSleetThunderSun: 'stormy',
  HeavySleetThunderSun: 'stormy',
  LightSnowThunderSun: 'stormy',
  HeavySnowThunderSun: 'stormy',
  DrizzleThunder: 'stormy',
  LightSleetThunder: 'stormy',
  HeavySleetThunder: 'stormy',
  LightSnowThunder: 'stormy',
  HeavySnowThunder: 'stormy',
  DrizzleSun: 'rainy',
  RainSun: 'rainy',
  LightSleetSun: 'cloudy',
  HeavySleetSun: 'cloudy',
  LightSnowSun: 'snowy',
  HeavysnowSun: 'snowy',
  Drizzle: 'rainy',
  LightSleet: 'cloudy',
  HeavySleet: 'cloudy',
  LightSnow: 'snowy',
  HeavySnow: 'snowy'
}

const textMap = {
  Sun: 'Sunny',
  LightCloud: 'Light clouds',
  PartlyCloud: 'Partly cloudy',
  Cloud: 'Cloudy',
  LightRainSun: 'Light rain',
  LightRainThunderSun: 'Light rain, sun and thunder',
  SleetSun: 'Sleet and sun',
  SnowSun: 'Snow and sun',
  LightRain: 'Light rain',
  Rain: 'Rain',
  RainThunder: 'Rain and thunder',
  Sleet: 'Sleet',
  Snow: 'Snow',
  SnowThunder: 'Snow and thunder',
  Fog: 'Foggy',
  SleetSunThunder: 'Sleet, sun and thunder',
  SnowSunThunder: 'Snow, sun and thunder',
  LightRainThunder: 'Light rain and thunder',
  SleetThunder: 'Sleet and thunder',
  DrizzleThunderSun: 'Drizzle, sun and thunder',
  RainThunderSun: 'Rain, sun and thunder',
  LightSleetThunderSun: 'Light sleet, sun and thunder',
  HeavySleetThunderSun: 'Heavy sleet, sun and thunder',
  LightSnowThunderSun: 'Light snow,  sun and thunder',
  HeavySnowThunderSun: 'Heavy snow, sun and thunder',
  DrizzleThunder: 'Drizzle and thunder',
  LightSleetThunder: 'Light sleet and thunder',
  HeavySleetThunder: 'Heavy sleet and thunder',
  LightSnowThunder: 'Light snow and thunder',
  HeavySnowThunder: 'Heavy snow and thunder',
  DrizzleSun: 'Drizzle and sun',
  RainSun: 'Rain and sun',
  LightSleetSun: 'Light sleet and sun',
  HeavySleetSun: 'Heavy sleet and sun',
  LightSnowSun: 'Light snow and sun',
  HeavysnowSun: 'Heavy snow and sun',
  Drizzle: 'Drizzle',
  LightSleet: 'Light sleet',
  HeavySleet: 'Heavy sleet',
  LightSnow: 'Light snow',
  HeavySnow: 'Heavy snow'
}

const unitMap = {
  celsius: 'C',
  farenheight: 'F',
  kelvin: 'K'
}

const levelToNumber = level => {
  switch (level) {
    case 'i.u.': // Ingen uppgift
      return null
    case 'i.h.': // Inga halter
      return 0
    case 'L': // Låga halter
      return 1
    case 'L-M': // Låga-måttliga halter
      return 2
    case 'M': // Måttliga halter
      return 3
    case 'M-H': // Måttliga-höga halter
      return 4
    case 'H': // Höga halter
      return 5
    case 'H-H+': // Mycket höga halter
      return 6
    default:
      if (parseInt(level, 10) === level) {
        return level === 0 ? level : level + 1
      }
      console.log('unknown level', level)
      return 0
  }
}

const levelToColor = level => {
  switch (level) {
    case null:
      return null
    case 0: // Inga halter
      return null
    case 1: // Låga halter
      return 'skyblue'
    case 2: // Låga-måttliga halter
      return 'lightgreen'
    case 3: // Måttliga halter
      return 'white'
    case 4: // Måttliga-höga halter
      return 'yellow'
    case 5: // Höga halter
      return 'orange'
    case 6: // Mycket höga halter
      return 'red'
    default:
      console.log('unknown color', level)
      return 0
  }
}

class Forecast extends React.Component {
  componentDidMount() {
    const { dispatch, config } = this.props

    this.forecastInterval = setInterval(
      () => dispatch(loadForecast(config)),
      60 * 60 * 1000
    )
    dispatch(loadForecast(config))
  }

  render() {
    const { forecasts, pollen } = this.props
    return (
      <React.Fragment>
        {forecasts &&
          forecasts
            .filter(
              forecast => moment().diff(moment(forecast.from), 'date') < 0
            )
            .map(forecast => (
              <div className="forecast" key={forecast.from}>
                <strong>
                  <Moment calendar={calendarStrings}>{forecast.from}</Moment>
                </strong>
                <div className="weather">
                  <div className="weathericon">
                    <div className={iconMap[forecast.icon]} />
                  </div>
                  <div className="weathervalues">
                    {forecast.temperature && (
                      <div>
                        {forecast.temperature.value}&deg;
                        {unitMap[forecast.temperature.unit]}{' '}
                        {textMap[forecast.icon]}
                      </div>
                    )}
                    {forecast.rain && <small>{forecast.rain}</small>}
                    {forecast.windSpeed && (
                      <small>{forecast.windSpeed.mps} m/s</small>
                    )}
                    {forecast.pressure && (
                      <small>
                        {forecast.pressure.value} {forecast.pressure.unit}
                      </small>
                    )}
                  </div>
                  {pollen.get(moment(forecast.from).format('YYYY-MM-DD')) && (
                    <div className="pollen">
                      {pollen
                        .get(moment(forecast.from).format('YYYY-MM-DD'))
                        .map((values, source) =>
                          values
                            .map((data, type) =>
                              data.set(
                                'level',
                                levelToNumber(data.get('value'))
                              )
                            )
                            .map((data, type) =>
                              data.set('color', levelToColor(data.get('level')))
                            )
                        )
                        .map((values, source) => (
                          <div>
                            <div>
                              <strong>{source}</strong>
                            </div>
                            {values
                              .filter(data => data.get('level') > 0)
                              .sort((a, b) => b.get('level') - a.get('level'))
                              .map((data, type) => (
                                <div
                                  style={{
                                    fontSize: 10 + data.get('level') * 8,
                                    color: data.get('color')
                                  }}>
                                  {type}
                                </div>
                              ))
                              .toArray()}
                          </div>
                        ))
                        .toArray()}
                    </div>
                  )}
                </div>
              </div>
            ))}
      </React.Fragment>
    )
  }
}

//<div className="cloudy"></div>
//<div className="rainy"></div>
//<div className="snowy"></div>
//<div className="rainbow"></div>
//<div className="starry"></div>
//<div className="stormy"></div>

const mapStateToProps = state => ({
  forecasts: state.getIn(['forecast', 'forecast']),
  pollen: state.getIn(['forecast', 'pollen']),
  current: state.getIn(['forecast', 'current'])
})

export default connect(mapStateToProps)(Forecast)
