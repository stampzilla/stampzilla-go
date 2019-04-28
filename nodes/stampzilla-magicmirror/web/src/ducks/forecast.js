import yrno from 'yr.no-forecast'
import moment from 'moment'
import { Map } from 'immutable'
import axios from 'axios'

export const LOADING = 'forecast/LOADING'
export const SUCCESS = 'forecast/SUCCESS'
export const SUCCESS_POLLENKOLL = 'forecast/SUCCESS_POLLENKOLL'
export const SUCCESS_BREEZOMETER = 'forecast/SUCCESS_BREEZOMETER'

const initialState = Map({
  forecast: null,
  current: null,
  loading: false,
  pollen: Map()
})

export const loadForecast = config => {
  return dispatch => {
    dispatch({
      type: LOADING
    })

    const LOCATION = {
      lat: 56.870349,
      lon: 14.541664
    }

    const services = []

    services.push(
      yrno()
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
    )

    if (config.hasIn(['breezometer', 'pollen', 'url'])) {
      services.push(
        axios
          .get(
            'proxy?url='.concat(
              encodeURIComponent(
                `${config.getIn(['breezometer', 'pollen', 'url'])}&lat=${
                  LOCATION.lat
                }&lon=${LOCATION.lon}`
              )
            )
          )
          .then(resp => {
            dispatch({
              type: SUCCESS_BREEZOMETER,
              data: resp.data,
              config: config.getIn(['breezometer', 'pollen'])
            })
          })
          .catch(err => {
            console.error(err)
            return false
          })
      )
    }

    if (config.hasIn(['pollenkoll', 'url'])) {
      services.push(
        axios
          .get(
            'proxy?url='.concat(
              encodeURIComponent(config.getIn(['pollenkoll', 'url']))
            )
          )
          .then(resp => {
            dispatch({
              type: SUCCESS_POLLENKOLL,
              data: resp.data,
              config: config.get('pollenkoll')
            })
          })
          .catch(() => {
            return false
          })
      )
    }

    return services
  }
}

export default (state = initialState, action) => {
  const { data, config } = action
  const values = {}

  switch (action.type) {
    case LOADING:
      return state.set('loading', true)
    case SUCCESS:
      return state
        .set('forecast', action.forecast)
        .set('current', action.current)
    case SUCCESS_POLLENKOLL:
      const cities = data.filter(
        row => config.get('city').indexOf(row.name) !== -1
      )

      cities.forEach(city => {
        const types = {}
        city.pollen.forEach(row => {
          const type = row.type
          Object.keys(row).forEach(key => {
            if (key === 'type') {
              return
            }

            const parts = key.match(/day([0-9])_(.+)/)

            if (types[type] === undefined) {
              types[type] = {}
            }

            if (types[type][parts[1]] === undefined) {
              types[type][parts[1]] = {}
            }
            types[type][parts[1]][parts[2]] = row[key]
          })
        })

        Object.keys(types).forEach(type => {
          Object.values(types[type]).forEach(row => {
            if (values[row.date] === undefined) {
              values[row.date] = {}
            }

            if (values[row.date][city.name] === undefined) {
              values[row.date][city.name] = {}
            }

            values[row.date][city.name][type] = {
              desc: row.desc,
              value: row.value
            }
          })
        })
      })

      return state.set('pollen', state.get('pollen').mergeDeep(values))
    case SUCCESS_BREEZOMETER:
      Object.values(data.data).forEach(row => {
        const { date, plants } = row

        if (plants) {
          Object.values(plants).forEach(plant => {
            if (!plant.data_available) {
              return
            }

            if (values[date] === undefined) {
              values[date] = {}
            }
            if (values[date]['BreezoMeter'] === undefined) {
              values[date]['BreezoMeter'] = {}
            }
            values[date]['BreezoMeter'][plant.display_name] = {
              value: plant.index.value,
              desc: plant.index.category
            }
          })
        }
      })

      return state.set('pollen', state.get('pollen').mergeDeep(values))
    default:
      return state
  }
}
