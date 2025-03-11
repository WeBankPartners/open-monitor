<template>
  <Badge :count="filterParamsCount" class="mr-2">
    <Dropdown
      trigger="click"
      transfer
      transfer-class-name='search-badge-dropdown'
      @on-visible-change='onDropdownVisibleChange'
    >
      <div class="badge-content">
        <Icon type="md-search" class="search-icon mr-1" />
        {{$t('m_filter')}}
      </div>
      <template slot='list'>
        <Form ref="fliters" :label-width="70" class="drop-down-content" @click="(e) => {e.stopPropagation()}">
          <FormItem :label="$t('m_alarm_query')">
            <Input v-model.trim="filters.query"
                   clearable
                   :placeholder="$t('m_alarm_query_tips')"
                   @on-change="onFilterChange"
            />
          </FormItem>
          <FormItem :label="$t('m_alarm_level')">
            <Select
              v-if='isSelectShow'
              v-model="filters.priority"
              placement='bottom'
              multiple
              transfer
              filterable
              :placeholder="$t('m_please_select') + $t('m_alarm_level')"
              @on-change="onFilterChange"
            >
              <Option v-for="item in filtersPriorityOptions" :label="item.name" :value="item.value" :key="item.value">
                {{item.name}}
              </Option>
            </Select>
          </FormItem>

          <FormItem :label="$t('m_alarmName')">
            <Select
              v-if='isSelectShow'
              v-model="filters.alarm_name"
              multiple
              filterable
              transfer
              placement='bottom'
              :placeholder="$t('m_please_select') + $t('m_alarmName')"
              @on-change="onFilterChange"
              @on-query-change="(query) => {
                if (query) {
                  filterParams.alarmName = query
                  filterParams.endpoint = ''
                  filterParams.metric = ''
                  onFilterOptions()
                }
              }"
            >
              <Option v-for="(name, index) in filtersAlarmNameOptions" :label="name" :value="name" :key="index">
                {{name}}
              </Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_metric')">
            <Select
              v-if='isSelectShow'
              v-model="filters.metric"
              multiple
              filterable
              transfer
              placement='bottom'
              :placeholder="$t('m_please_select') + $t('m_metric')"
              @on-change="onFilterChange"
              @on-query-change="(query) => {
                if (query) {
                  filterParams.metric = query
                  filterParams.alarmName = ''
                  filterParams.endpoint = ''
                  onFilterOptions()
                }
              }"
            >
              <Option v-for="(name, index) in filtersMetricOptions" :label="name" :value="name" :key="index">
                {{name}}
              </Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_endpoint')">
            <Select
              v-if='isSelectShow'
              v-model="filters.endpoint"
              multiple
              filterable
              transfer
              placement='bottom'
              :placeholder="$t('m_please_select') + $t('m_endpoint')"
              @on-change="onFilterChange"
              @on-query-change="(query) => {
                if (query) {
                  filterParams.endpoint = query
                  filterParams.metric = ''
                  filterParams.alarmName = ''
                  onFilterOptions()
                }
              }"
            >
              <Option v-for="(item, index) in filtersEndpointOptions" :label="item.displayName" :value="item.displayName + '$*$' + item.name" :key="index">
                {{item.displayName}}
              </Option>
            </Select>
          </FormItem>
          <Button type="primary" @click.stop="onResetButtonClick">{{$t('m_reset')}}</Button>
        </Form>
      </template>
    </Dropdown>
  </Badge>

</template>

<script>
import {
  debounce, cloneDeep, isEmpty, find
} from 'lodash'

const initFilters = {
  query: '',
  alarm_name: [],
  metric: [],
  endpoint: [],
  priority: []
}

const initFilterParams = {
  status: '',
  alarmName: '',
  endpoint: '',
  metric: ''
}

export default ({
  props: {
    tempFilters: String,
  },
  watch: {
    tempFilters: {
      handler(newVal) {
        if (newVal && newVal !== JSON.stringify(this.filters)) {
          this.filters = JSON.parse(newVal)
          this.onFilterChange()
        }
      }
    }
  },
  data() {
    return {
      filtersAlarmNameOptions: [],
      filtersEndpointOptions: [],
      filtersMetricOptions: [],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      filters: cloneDeep(initFilters),
      filtersPriorityOptions: [
        {
          name: this.$t('m_low'),
          value: 'low'
        },
        {
          name: this.$t('m_medium'),
          value: 'medium'
        },
        {
          name: this.$t('m_high'),
          value: 'high'
        }
      ],
      filterParams: cloneDeep(initFilterParams),
      isSelectShow: false
    }
  },
  computed: {
    filterParamsCount() {
      let count = 0
      for (const key in this.filters) {
        if (this.filters[key] && this.filters[key].length) {
          count++
        }
      }
      return count
    }
  },
  mounted(){
    document.querySelector('.drop-down-content.ivu-form.ivu-form-label-right').addEventListener('click', e => e.stopPropagation())
    document.querySelector('.badge-content').addEventListener('click', () => {
      this.filterParams = cloneDeep(initFilterParams)
      this.filterParams.status = this.$route.path === '/alarmManagement' ? 'firing' : ''
      this.getFilterAllOptions()
    })
  },
  methods: {
    onFilterOptions: debounce(function () {
      this.getFilterAllOptions()
    }, 800),
    getFilterAllOptions() {
      this.request('POST', this.apiCenter.alarmProblemOptions, this.filterParams, res => {
        this.filtersAlarmNameOptions = res.alarmNameList
        this.filtersEndpointOptions = res.endpointList
        this.filtersMetricOptions = res.metricList
        this.processOptions()
      })
    },
    onResetButtonClick() {
      this.filters = cloneDeep(initFilters)
      this.onFilterChange()
    },
    limitFiltersLength() {
      for (const key in this.filters) {
        if (Array.isArray(this.filters[key]) && this.filters[key].length > 3) {
          this.filters[key].splice(3)
        }
      }
    },
    processOptions() {
      !isEmpty(this.filters.alarm_name) && this.filters.alarm_name.forEach(name => {
        if (!this.filtersAlarmNameOptions.includes(name)) {
          this.filtersAlarmNameOptions.unshift(name)
        }
      })

      !isEmpty(this.filters.metric) && this.filters.metric.forEach(val => {
        if (!this.filtersMetricOptions.includes(val)) {
          this.filtersMetricOptions.unshift(val)
        }
      })

      !isEmpty(this.filters.endpoint) && this.filters.endpoint.forEach(val => {
        if (!find(this.filtersEndpointOptions, {
          displayName: val.split('$*$')[0],
          name: val.split('$*$')[1]
        })) {
          this.filtersEndpointOptions.unshift({
            displayName: val.split('$*$')[0],
            name: val.split('$*$')[1]
          })
        }
      })

    },
    onFilterChange: debounce(function (){
      this.limitFiltersLength()
      this.processOptions()
      this.onFilterConditionChange()
    }, 400),
    onFilterConditionChange() {
      this.$emit('filtersChange', cloneDeep(this.filters))
    },
    onDropdownVisibleChange(show) {
      this.isSelectShow = show
    }
  }
})
</script>

<style lang='less'>
.search-badge-dropdown {
  min-height: 370px
}
</style>

<style scoped lang='less'>
.badge-content {
  display: flex;
  justify-content: center;
  align-items: center;
  min-width: 60px;
  font-size: 14px;
  color: #5981ef;
  background-color: #f8f8f8;
  padding: 6px;
  border-radius: 10%;
  cursor: pointer;
}
.drop-down-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 400px
}
.drop-down-content > div {
  width: 80%;
  margin: 10px auto
}

</style>
