<template>
  <Badge :count="filterParamsCount" class="mr-2">
    <Dropdown trigger="click">
      <div class="badge-content">
        <Icon type="md-search" class="search-icon mr-1" />
        {{$t('m_filter')}}
      </div>
      <template #list>
        <Form ref="fliters" :label-width="70" class="drop-down-content" @click="(e) => {e.stopPropagation()}">
          <FormItem :label="$t('m_alarmName')">
            <Select
              v-model="filters.alarm_name"
              multiple
              filterable
              :placeholder="$t('m_please_select') + $t('m_alarmName')"
              @on-change="onFilterConditionChange"
            >
              <Option v-for="name in filtersAlarmNameOptions" :value="name" :key="name">
                {{name}}
              </Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_metric')">
            <Select
              v-model="filters.metric"
              multiple
              filterable
              :placeholder="$t('m_please_select') + $t('m_metric')"
              @on-change="onFilterConditionChange"
            >
              <Option v-for="name in filtersMetricOptions" :value="name" :key="name">
                {{name}}
              </Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_endpoint')">
            <Select
              v-model="filters.endpoint"
              multiple
              filterable
              :placeholder="$t('m_please_select') + $t('m_endpoint')"
              @on-change="onFilterConditionChange"
            >
              <Option v-for="name in filtersEndpointOptions" :value="name" :key="name">
                {{name}}
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
import debounce from 'lodash/debounce'
import cloneDeep from 'lodash/cloneDeep'

const initFilters = {
  alarm_name: [],
  metric: [],
  endpoint: []
}

export default ({
  props: {
    tempFilters: String,
  },
  watch: {
    tempFilters: {
      handler(newVal) {
        if (newVal) {
          this.filters = JSON.parse(newVal)
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
      filters: cloneDeep(initFilters)
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
    this.getFilterAllOptions()
    document.querySelector('.drop-down-content.ivu-form.ivu-form-label-right').addEventListener('click', e => e.stopPropagation())
  },
  methods: {
    getFilterAllOptions() {
      const api = '/monitor/api/v1/alarm/problem/options'
      this.request('GET', api, {}, res => {
        this.filtersAlarmNameOptions = res.alarmNameList
        this.filtersEndpointOptions = res.endpointList
        this.filtersMetricOptions = res.metricList
      })
    },
    onFilterConditionChange: debounce(function () {
      this.$emit('filtersChange', cloneDeep(this.filters))
    }, 300),
    onResetButtonClick() {
      this.filters = cloneDeep(initFilters)
      this.onFilterConditionChange()
    }
  }

})
</script>

<style scoped lang='less'>
.badge-content {
  display: flex;
  justify-content: center;
  align-items: center;
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
