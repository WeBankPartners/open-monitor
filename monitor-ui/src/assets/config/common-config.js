export const dataPick = [
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
  },
  {
    value: -21600,
    label: '6h'
  },
  {
    value: -43200,
    label: '12h'
  },
];

export const autoRefreshConfig = [
  {
    value: -1,
    label: "off"
  },
  {
    value: 5,
    label: '5s'
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
  {label: 'm_high', value: 'high'},
  {label: 'm_medium', value: 'medium'},
  {label: 'm_low', value: 'low'}
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
  'cyan', 
  'geekblue', 
  'purple']

export const colorList = {
  'cyan': '#13c2c2',
  'blue': '#1890ff',
  'geekblue': '#2f54eb',
  'purple': '#722ed1',
  'primary': '#2d8cf0',
  'success': '#19be6b',
  'warning': '#f90',
  'green': '#b7eb8f',
  'red': '#52c41a',
  'yellow': '#fadb14',
  'pink': '#ffadd2',
  'magenta': '#eb2f96',
  'volcano': '#fa541c',
  'orange': '#e8c16d'
}
export const cycleOption = [
  {label: 'm_all', value: 'All'},
  {label: 'm_monday', value: 'Monday'},
  {label: 'm_tuesday', value: 'Tuesday'},
  {label: 'm_wednesday', value: 'Wednesday'},
  {label: 'm_thursday', value: 'Thursday'},
  {label: 'm_friday', value: 'Friday'},
  {label: 'm_saturday', value: 'Saturday'},
  {label: 'm_sunday', value: 'Sunday'}
]

export const collectionInterval =[
  {label: '1s', value: 1},
  {label: '5s', value: 5},
  {label: '10s', value: 10},
  {label: '30s', value: 30},
  {label: '60s', value: 60},
]