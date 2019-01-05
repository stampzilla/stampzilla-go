import React, { Component } from 'react';
import { push } from 'connected-react-router';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import moment from 'moment';
import Moment from 'react-moment';
import { loadForecast } from '../../ducks/forecast';

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
  HeavySnow: 'snowy',
};

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
  HeavySnow: 'Heavy snow',
};

const unitMap = {
  celsius: 'C',
  farenheight: 'F',
  kelvin: 'K',
};

class Home extends Component {
  constructor(props) {
    super(props);
    this.clock = React.createRef();
    this.seconds = React.createRef();
    this.date = React.createRef();
  }

  componentDidMount() {
    const { dispatch } = this.props;

    const updateClock = () => {
      this.clock.current.innerHTML = moment().format('HH:mm');
      this.seconds.current.innerHTML = moment().format('ss');
      this.date.current.innerHTML = moment().format('dddd - D MMMM');
    }
    updateClock();
    setInterval(updateClock, 1000);

    this.forecastInterval = setInterval(() => dispatch(loadForecast()), 60*60*1000);
    dispatch(loadForecast());
  }

  componentWillUnmount() {
    clearTimeout(this.forecastInterval);
  }

  render() {
    const { forecasts } = this.props; 

    const calendarStrings = {
        lastDay : '[Yesterday]',
        sameDay : '[Today]',
        nextDay : '[Tomorrow]',
        lastWeek : '[last] dddd [at] LT',
        nextWeek : 'dddd',
        sameElse : 'L'
    };


    return (
      <div>
        <div className="clock">
          <div ref={this.clock}>00:00</div>
          <div className="seconds" ref={this.seconds}>00</div>
        </div>
        <div className="date" ref={this.date}>
        -  
        </div>

        {forecasts && forecasts.filter(forecast => moment().diff(moment(forecast.from), 'date') < 0).map(forecast => (
          <div className="forecast">
            <strong><Moment calendar={calendarStrings}>{forecast.from}</Moment></strong>
            <div className="weather">
              <div className="weathericon">
                <div className={iconMap[forecast.icon]}></div>
              </div>
              <div className="weathervalues">
                {forecast.temperature && 
                  <div >{forecast.temperature.value}&deg;{unitMap[forecast.temperature.unit]} {textMap[forecast.icon]}</div>
                }
                {forecast.rain && 
                  <small>{forecast.rain}</small>
                }
                {forecast.windSpeed && 
                  <small>{forecast.windSpeed.mps} m/s</small>
                }
                {forecast.pressure && 
                  <small>{forecast.pressure.value} {forecast.pressure.unit}</small>
                }
              </div>
            </div>
          </div>
        ))}


        <div style={{ position: 'relative' }}>
        </div>

      </div>
    );
  }
}

          //<div className="cloudy"></div>
          //<div className="rainy"></div>
          //<div className="snowy"></div>
          //<div className="rainbow"></div>
          //<div className="starry"></div>
          //<div className="stormy"></div>

const mapStateToProps = ({ forecast }) => ({
  forecasts: forecast.forecast,
  current: forecast.current,
})

export default connect(mapStateToProps)(Home)
