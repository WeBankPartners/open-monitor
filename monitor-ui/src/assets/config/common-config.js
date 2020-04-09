export const dataPick = [
  {
    value: -1,
    label: "off"
  },
  {
    value: -1800,
    label: '30m'
  },
  {
    value: -3600,
    label: '1h'
  },
  {
    value: -10800,
    label: '3h'
  }
];

export const autoRefreshConfig = [
  {
    value: -1,
    label: "off"
  },
  {
    value: 10,
    label: '10s'
  },
  {
    value: 30,
    label: '30s'
  },
  {
    value: 60,
    label: '1m'
  },
  {
    value: 300,
    label: '5m'
  }
]

export const thresholdList = [
  {label: '>', value: '>'},
  {label: '>=', value: '>='},
  {label: '<', value: '<'},
  {label: '<=', value: '<='},
  {label: '==', value: '=='},
  {label: '!=', value: '!='}
]

export const lastList = [
  {label: 'sec', value: 's'},
  {label: 'min', value: 'm'},
  {label: 'hour', value: 'h'}
]

export const priorityList = [
  {label: 'high', value: 'high'},
  {label: 'medium', value: 'medium'},
  {label: 'low', value: 'low'}
]

export const endpointTag = {
  host: 'cyan',
  mysql: 'blue',
  redis: 'geekblue',
  tomcat: 'purple'
}

export const randomColor = [
  'primary', 
  'success', 
  'warning', 
  'blue', 
  'green', 
  'red', 
  'yellow', 
  'pink', 
  'magenta', 
  'volcano', 
  'orange', 
  'gold', 
  'lime', 
  'cyan', 
  'geekblue', 
  'purple']