import yrno from 'yr.no-forecast'
import moment from 'moment'
import { List, Map } from 'immutable'

export const LOADING = 'forecast/LOADING'
export const SUCCESS = 'forecast/SUCCESS'

const initialState = Map({
  forecast: null,
  current: null,
  loading: false
})

export const loadForecast = () => {
  return dispatch => {
    dispatch({
      type: LOADING
    })

    const LOCATION = {
      lat: 56.870349,
      lon: 14.541664
    }

    return yrno()
      .getWeather(LOCATION)
      .then(weather => {
        Promise.all([
          weather.getFiveDaySummary(),
          weather.getForecastForTime(moment().startOf('day'))
        ]).then(result => {
          dispatch({
            type: SUCCESS,
            forecast: result[0],
            current: result[1]
          })
        })
      })
  }
}

export default (state = initialState, action) => {
  switch (action.type) {
    case LOADING:
      return state.set('loading', true)
    case SUCCESS:
      return state
        .set('forecast', action.forecast)
        .set('current', action.current)
    default:
      return state
  }
}
